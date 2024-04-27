package mvp6_4

import (
	"errors"
	"fmt"

	"github.com/teivah/broadcast"
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

// TODO Why is 100 blocking?
const maxCacheControllers = 1000

type msiState = int32

const (
	invalid msiState = iota // Default has to be invalid
	shared
	modified
)

type busRequestEvent struct {
	id int
	// Mutually exclusive
	read       bool
	write      bool
	invalidate bool
	addr       int32
	response   *broadcast.Relay[busResponseEvent]
}

type busResponseEvent struct {
	// TODO To delete?
	pendingWrite bool
	memoryChange *instructionForNextInvalidate
}

type cacheController struct {
	ctx *risc.Context
	id  int
	mmu *memoryManagementUnit
	// msi implements the MSI protocol.
	// Source: https://www.youtube.com/watch?v=gAUVAel-2Fg.
	msi map[int32]msiState
	co.Coroutine[ccReq, ccResp]
	l1d                           *comp.LRUCache
	bus                           *comp.Broadcast[busRequestEvent]
	instructionsForNextInvalidate map[int32]instructionForNextInvalidate
}

// TODO Better name
type instructionForNextInvalidate struct {
	addrs  []int32
	memory []int8
}

func newCacheController(id int, ctx *risc.Context, mmu *memoryManagementUnit, bus *comp.Broadcast[busRequestEvent]) *cacheController {
	cc := &cacheController{
		ctx:                           ctx,
		id:                            id,
		mmu:                           mmu,
		msi:                           make(map[int32]msiState),
		l1d:                           comp.NewLRUCache(l1DCacheLineSize, l1DCacheSize),
		bus:                           bus,
		instructionsForNextInvalidate: make(map[int32]instructionForNextInvalidate),
	}
	cc.Coroutine = co.New(cc.coRead)
	return cc
}

type ccReq struct {
	addrs []int32
}

type ccResp struct {
	memory []int8
	done   bool
}

func (cc *cacheController) snoop() bool {
	for _, evt := range cc.bus.Read(cc.id) {
		event := evt.Data

		// Discard events from the same cache controller ID
		if event.id == cc.id {
			evt.Commit()
			continue
		}

		state := cc.msi[event.addr]
		if event.read {
			switch state {
			case invalid:
			case modified:
				if _, exists := cc.ctx.PendingWriteMemoryIntention[event.addr]; exists {
					// A change is not written in L1 yet
					event.response.Notify(busResponseEvent{true, nil})
					continue
				}

				if _, exists := cc.ctx.PendingWriteMemoryIntention[event.addr]; exists {
					// The change is not written in L1 yet
					event.response.Notify(busResponseEvent{true, nil})
					continue
				}

				// Write-back
				memory, exists := cc.l1d.GetCacheLine(event.addr)
				if !exists {
					panic("memory address should exist")
				}
				// TODO Time to write
				cc.mmu.writeToMemory(event.addr, memory)
				cc.msi[event.addr] = shared
				fmt.Println(cc.id, "snoop", "read", "modified to shared", event.addr, memory)
			case shared:
			}
			evt.Commit()
		} else if event.write {
			if _, exists := cc.ctx.PendingWriteMemoryIntention[event.addr]; exists {
				// A change is not written in L1 yet
				event.response.Notify(busResponseEvent{true, nil})
				continue
			}

			switch state {
			case invalid:
			case modified:
				// Write-back
				// TODO Time to write
				memory, exists := cc.l1d.EvictCacheLine(event.addr)
				if !exists {
					panic("memory address should exist")
				}
				cc.mmu.writeToMemory(event.addr, memory)
				cc.msi[event.addr] = invalid
				fmt.Println(cc.id, "snoop", "write", "modified to invalid", event.addr, memory)
			case shared:
				_, _ = cc.l1d.EvictCacheLine(event.addr)
				// TODO Really?
				//if !exists {
				//	panic("memory address should exist")
				//}
				cc.msi[event.addr] = invalid
				fmt.Println(cc.id, "snoop", "write", "shared to invalid", event.addr)
			}
			evt.Commit()
		} else if event.invalidate {
			// TODO I should see the invalidation request at delta 32
			if v, exists := cc.instructionsForNextInvalidate[event.addr]; exists {
				event.response.Notify(busResponseEvent{false, &v})
				delete(cc.instructionsForNextInvalidate, event.addr)
			}
			switch state {
			case invalid:
				// Nothing
			case modified:
				// Write-back
				// TODO Time to write
				memory, exists := cc.l1d.EvictCacheLine(event.addr)
				if !exists {
					panic("memory address should exist")
				}
				cc.mmu.writeToMemory(event.addr, memory)
				cc.msi[event.addr] = invalid
			case shared:
				cc.l1d.EvictCacheLine(event.addr)
				cc.msi[event.addr] = invalid
			}
			evt.Commit()
		} else {
			panic("unknown event type")
		}
	}

	return false
}

func (cc *cacheController) coRead(r ccReq) ccResp {
	//if id, exists := cc.ctx.PendingWriteMemoryIntention[getAlignedMemoryAddress(r.addrs)]; exists && id != cc.id {
	//	// A write is pending
	//	return ccResp{}
	//}

	exists := cc.isAddressInL1(r.addrs)
	if exists {
		memory := cc.getFromL1(r.addrs)
		// TODO Check it executes in 3 cycles
		cycles := latency.L1Access
		return cc.ExecuteWithCheckpoint(r, func(r ccReq) ccResp {
			if cycles > 0 {
				cycles--
				return ccResp{}
			}
			cc.Reset()
			return ccResp{
				memory: memory,
				done:   true,
			}
		})
	}

	return cc.ExecuteWithCheckpoint(r, cc.coTryRead)
}

func (cc *cacheController) coTryRead(r ccReq) ccResp {
	_, listener := cc.msiRead(r.addrs)
	cc.Checkpoint(func(r ccReq) ccResp {
		if listener != nil {
			// In this case, we may have to wait for a write-back
			pending := false
		loop:
			for {
				select {
				// TODO Close?
				case resp := <-listener.Ch():
					pending = pending || resp.pendingWrite
				default:
					break loop
				}
			}
			if pending {
				return ccResp{}
			}
		}

		cycles := latency.MemoryAccess
		// TODO Shouldn't we read from L1?
		addr, line := cc.mmu.fetchCacheLine(r.addrs[0], l1DCacheLineSize)
		return cc.ExecuteWithCheckpoint(r, func(r ccReq) ccResp {
			if cycles > 0 {
				cycles--
				return ccResp{}
			}

			if cc.msi[addr] != shared {
				// Cache line is stale (another core modified it)
				// We have to read again
				fmt.Println(cc.id, "stale", addr)
				return cc.ExecuteWithCheckpoint(r, cc.coTryRead)
			}

			cc.Reset()
			cc.pushLineToL1(addr, line)
			data := cc.getFromL1(r.addrs)

			return ccResp{
				memory: data,
				done:   true,
			}
		})
	})
	return ccResp{}
}

func (cc *cacheController) pushLineToL1(addr int32, line []int8) {
	evicted := cc.l1d.PushLine(addr, line)
	if len(evicted) == 0 {
		return
	}
	// TODO Delay
	cc.mmu.writeToMemory(addr, line)
}

func getAlignedMemoryAddress(addrs []int32) int32 {
	addr := addrs[0]
	return addr - (addr % l1DCacheLineSize)
}

func (cc *cacheController) msiRead(addrs []int32) (int32, *broadcast.Listener[busResponseEvent]) {
	addr := getAlignedMemoryAddress(addrs)
	state := cc.msi[addr]
	switch state {
	case invalid:
		// TODO We may have to wait for write-back
		cc.msi[addr] = shared
		return addr, cc.readRequest(addr)
	case modified:
		// Nothing
		return addr, nil
	case shared:
		// Nothing
		return addr, nil
	default:
		panic("unknown state")
	}
}

type invalidationEvent struct {
	addr int32
}

func (cc *cacheController) msiWrite(addrs []int32, invalidations map[int32]bool) (int32, *broadcast.Listener[busResponseEvent], *invalidationEvent, error) {
	addr := getAlignedMemoryAddress(addrs)
	state := cc.msi[addr]
	switch state {
	case invalid:
		// TODO We may have to wait for write-back
		cc.msi[addr] = modified
		return addr, cc.writeRequest(addr), nil, nil
	case modified:
		// Nothing
		return addr, nil, nil, nil
	case shared:
		if invalidations[addr] {
			// Only one cache controller can emit an invalidation event during the
			// same cycle.
			cc.msi[addr] = invalid
			_, _ = cc.l1d.EvictCacheLine(addr)
			return addr, nil, nil, errors.New("multiple invalidation")
		}
		cc.msi[addr] = modified
		return addr, cc.invalidationRequest(addr), &invalidationEvent{addr}, nil
	default:
		panic("unknown state")
	}
}

func (cc *cacheController) readRequest(addr int32) *broadcast.Listener[busResponseEvent] {
	fmt.Println(cc.id, "read request", addr)
	response := broadcast.NewRelay[busResponseEvent]()
	cc.bus.Notify(busRequestEvent{
		id:       cc.id,
		read:     true,
		addr:     addr,
		response: response,
	})
	return response.Listener(maxCacheControllers)
}

func (cc *cacheController) writeRequest(addr int32) *broadcast.Listener[busResponseEvent] {
	fmt.Println(cc.id, "write request", addr)
	response := broadcast.NewRelay[busResponseEvent]()
	cc.bus.Notify(busRequestEvent{
		id:       cc.id,
		write:    true,
		addr:     addr,
		response: response,
	})
	return response.Listener(maxCacheControllers)
}

func (cc *cacheController) invalidationRequest(addr int32) *broadcast.Listener[busResponseEvent] {
	fmt.Println(cc.id, "invalidation request", addr)
	response := broadcast.NewRelay[busResponseEvent]()
	cc.bus.Notify(busRequestEvent{
		id:         cc.id,
		invalidate: true,
		addr:       addr,
		response:   response,
	})
	return response.Listener(maxCacheControllers)
}

func (cc *cacheController) isAddressInL1(addrs []int32) bool {
	_, exists := cc.l1d.Get(addrs[0])
	return exists
}

func (cc *cacheController) getFromL1(addrs []int32) []int8 {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		v, exists := cc.l1d.Get(addr)
		if !exists {
			panic("value presence should have been checked first")
		}
		memory = append(memory, v)
	}
	return memory
}

var delta = -1

func (cc *cacheController) writeToL1(addrs []int32, data []int8) {
	cc.l1d.Write(addrs[0], data)
	delta++
	fmt.Println(delta)
	fmt.Printf("write, id=%d, addr=%d, data=%d\n", cc.id, addrs[0], risc.I32FromBytes(data[0], data[1], data[2], data[3]))
	for _, line := range cc.l1d.Lines() {
		if line.Boundary[0] <= addrs[0] && addrs[0] < line.Boundary[1] {
			fmt.Printf("%v: ", line.Boundary)
			for i := 0; i < len(line.Data); i += 4 {
				fmt.Printf("%v ", line.Data[i])
			}
			fmt.Println()
		}
	}
	fmt.Println()
}

func (cc *cacheController) flush() int {
	additionalCycles := 0
	for _, line := range cc.l1d.Lines() {
		addr := line.Boundary[0]
		if cc.msi[addr] != modified {
			// If not modified, we don't write back the line in memory
			continue
		}

		additionalCycles += latency.MemoryAccess
		for i := 0; i < l3CacheLineSize; i++ {
			cc.mmu.writeToMemory(line.Boundary[0], line.Data)
		}
	}
	return additionalCycles
}

func (cc *cacheController) SetInstructionForInvalidateRequest(addr int32, addrs []int32, memory []int8) {
	cc.instructionsForNextInvalidate[addr] = instructionForNextInvalidate{
		addrs:  addrs,
		memory: memory,
	}
}
