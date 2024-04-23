package mvp6_4

import (
	"errors"
	"math"

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
	conflict     bool
	memoryChange *instructionForNextInvalidate
}

type cacheController struct {
	id  int
	mmu *memoryManagementUnit
	// msi implements the MSI protocol.
	// Source: https://www.youtube.com/watch?v=gAUVAel-2Fg.
	msi map[int32]msiState
	co.Coroutine[ccReq, ccResp]
	l1d                           *comp.LRUCache
	bus                           *broadcast.Relay[busRequestEvent]
	listener                      *broadcast.Listener[busRequestEvent]
	instructionsForNextInvalidate map[int32]instructionForNextInvalidate
}

// TODO Better name
type instructionForNextInvalidate struct {
	addrs  []int32
	memory []int8
}

func newCacheController(id int, mmu *memoryManagementUnit, bus *broadcast.Relay[busRequestEvent]) *cacheController {
	cc := &cacheController{
		id:                            id,
		mmu:                           mmu,
		msi:                           make(map[int32]msiState),
		l1d:                           comp.NewLRUCache(l1DCacheLineSize, l1DCacheSize),
		bus:                           bus,
		listener:                      bus.Listener(maxCacheControllers),
		instructionsForNextInvalidate: make(map[int32]instructionForNextInvalidate),
	}
	cc.Coroutine = co.New(cc.start)
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
	for {
		select {
		case event := <-cc.listener.Ch():
			// Discard events from the same cache controller ID
			if event.id == cc.id {
				continue
			}
			state := cc.msi[event.addr]
			if event.read {
				switch state {
				case invalid:
					// Nothing
				case modified:
					// Write-back
					// TODO Time to write
					memory, exists := cc.l1d.GetCacheLine(event.addr)
					if !exists {
						panic("memory address should exist")
					}
					cc.mmu.writeToMemory(event.addr, memory)
					cc.msi[event.addr] = shared
					// TODO true
					event.response.Notify(busResponseEvent{false, nil})
				case shared:
					// Nothing
				}
			} else if event.write {
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
					// TODO true
					event.response.Notify(busResponseEvent{false, nil})
				case shared:
					_, exists := cc.l1d.EvictCacheLine(event.addr)
					if !exists {
						panic("memory address should exist")
					}
					cc.msi[event.addr] = invalid
				}
			} else if event.invalidate {
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
			} else {
				panic("unknown event type")
			}
		default:
			return false
		}
	}
}

func (cc *cacheController) start(r ccReq) ccResp {
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

	// Read from memory
	alignedAddr, listener := cc.msiRead(r.addrs)
	cc.Checkpoint(func(r ccReq) ccResp {
		if cc.msi[alignedAddr] != shared {
			// Cache line isn't readable any more.
			// This can happen if there was an invalidation before having fetched the
			// cache line.
			cc.Reset()
			return ccResp{}
		}
		if listener != nil {
			// In this case, we may have to wait for a write-back
			conflict := false
		loop:
			for {
				select {
				// TODO Close?
				case resp := <-listener.Ch():
					conflict = conflict || resp.conflict
				default:
					break loop
				}
			}

			if conflict {
				// We have to wait
				return ccResp{}
			}
		}

		cycles := latency.MemoryAccess
		memory := cc.mmu.getFromMemory(r.addrs)
		return cc.ExecuteWithCheckpoint(r, func(r ccReq) ccResp {
			if cycles > 0 {
				cycles--
				return ccResp{}
			}
			cc.Reset()
			// TODO Delay
			addr, line := cc.mmu.fetchCacheLine(r.addrs[0], l1DCacheLineSize)
			cc.pushLineToL1(addr, line)

			return ccResp{
				memory: memory,
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

func (cc *cacheController) getAlignedMemoryAddress(addrs []int32) int32 {
	addr := addrs[0]
	return addr - (addr % l1DCacheLineSize)
}

func (cc *cacheController) msiRead(addrs []int32) (int32, *broadcast.Listener[busResponseEvent]) {
	addr := cc.getAlignedMemoryAddress(addrs)
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
	addr := cc.getAlignedMemoryAddress(addrs)
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

func (cc *cacheController) doesExecutionMemoryChangesExistsInL1(execution risc.Execution) bool {
	var minAddr int32 = math.MaxInt32
	for addr := range execution.MemoryChanges {
		minAddr = min(minAddr, addr)
	}
	return cc.isAddressInL1([]int32{minAddr})
}

func (cc *cacheController) writeToL1(addrs []int32, data []int8) {
	cc.l1d.Write(addrs[0], data)
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
