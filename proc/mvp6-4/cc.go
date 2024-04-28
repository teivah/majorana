package mvp6_4

import (
	"fmt"
	"slices"

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

type ccReadReq struct {
	addrs []int32
}

type ccReadResp struct {
	data []int8
	done bool
}

type ccWriteReq struct {
	cycle int
	addrs []int32
	data  []int8
}

type ccWriteResp struct {
	done bool
}

type cacheController struct {
	ctx       *risc.Context
	id        int
	mmu       *memoryManagementUnit
	l1d       *comp.LRUCache
	bus       *comp.Broadcast[busRequestEvent]
	msi       map[int32]msiState
	read      co.Coroutine[ccReadReq, ccReadResp]
	write     co.Coroutine[ccWriteReq, ccWriteResp]
	snoop     co.Coroutine[struct{}, bool]
	snoops    []co.Coroutine[struct{}, bool]
	snoopBusy map[int32]bool
	pending   map[int32]map[actionType]pendingState
}

type actionType int32

const (
	readAction actionType = iota
	writeAction
)

type pendingState int32

const (
	none pendingState = iota
	pendingL1Access
	pendingFetchCacheLineFromMemory
	pendingReadRequestResponse
	pendingWriteRequestResponse
	pendingInvalidationRequestResponse
)

func newCacheController(id int, ctx *risc.Context, mmu *memoryManagementUnit, bus *comp.Broadcast[busRequestEvent]) *cacheController {
	cc := &cacheController{
		ctx:       ctx,
		id:        id,
		mmu:       mmu,
		l1d:       comp.NewLRUCache(l1DCacheLineSize, l1DCacheSize),
		bus:       bus,
		msi:       make(map[int32]msiState),
		pending:   make(map[int32]map[actionType]pendingState),
		snoopBusy: make(map[int32]bool),
	}
	cc.read = co.New(cc.coRead)
	cc.read.Pre(func(r ccReadReq) bool {
		alignedAddr := getAlignedMemoryAddress(r.addrs)
		return cc.snoopBusy[alignedAddr]
	})
	cc.write = co.New(cc.coWrite)
	cc.write.Pre(func(r ccWriteReq) bool {
		alignedAddr := getAlignedMemoryAddress(r.addrs)
		return cc.snoopBusy[alignedAddr]
	})
	cc.snoop = co.New(cc.coSnoop)
	return cc
}

// Constraint: when adding a coroutine, the event has to be committed.
func (cc *cacheController) coSnoop(struct{}) bool {
	cc.snoops = slices.DeleteFunc(cc.snoops, func(snoop co.Coroutine[struct{}, bool]) bool {
		//return !snoop.Cycle(struct{}{})
		v := snoop.Cycle(struct{}{})
		if !v {
			return true
		}
		return false
	})

	for _, evt := range cc.bus.Read(cc.id) {
		event := evt.Data

		// Discard events from the same cache controller ID
		if event.id == cc.id {
			evt.Commit()
			continue
		}

		readPending, writePending := cc.getPendingIntention(event.alignedAddr)

		state := cc.msi[event.alignedAddr]
		if event.read {
			switch state {
			case invalid:
				// Nothing
				evt.Commit()
			case modified:
				if writePending != none {
					// Write is ongoing, we should wait before write-back
					event.response.Notify(busResponseEvent{true})
					continue
				}

				evt.Commit()
				// Write back
				cc.snoopBusy[event.alignedAddr] = true
				cycles := latency.MemoryAccess
				cc.snoops = append(cc.snoops, co.New(func(struct{}) bool {
					if cycles > 0 {
						cycles--
						event.response.Notify(busResponseEvent{true})
						return true
					}

					memory, exists := cc.l1d.GetCacheLine(event.alignedAddr)
					if !exists {
						panic("memory address should exist")
					}
					cc.mmu.writeToMemory(event.alignedAddr, memory)

					cc.msi[event.alignedAddr] = invalid
					cc.snoopBusy[event.alignedAddr] = false
					return false
				}))
			case shared:
				// Nothing
				evt.Commit()
			}
		} else if event.write {
			switch state {
			case invalid:
				// Nothing
				evt.Commit()
			case modified:
				if readPending != none || writePending != none {
					// Write is ongoing, we should wait before write-back
					event.response.Notify(busResponseEvent{true})
					continue
				}

				evt.Commit()
				// Write back
				cc.snoopBusy[event.alignedAddr] = true
				cycles := latency.MemoryAccess
				cc.snoops = append(cc.snoops, co.New(func(struct{}) bool {
					if cycles > 0 {
						cycles--
						event.response.Notify(busResponseEvent{true})
						return true
					}

					memory, exists := cc.l1d.GetCacheLine(event.alignedAddr)
					if !exists {
						panic("memory address should exist")
					}
					cc.mmu.writeToMemory(event.alignedAddr, memory)

					cc.msi[event.alignedAddr] = invalid
					cc.snoopBusy[event.alignedAddr] = false
					return false
				}))
			case shared:
				if readPending != none || writePending != none {
					continue
				}

				_, evicted := cc.l1d.EvictCacheLine(event.alignedAddr)
				if !evicted {
					panic("invalid state")
				}
				cc.msi[event.alignedAddr] = invalid
				evt.Commit()
			}
		} else if event.invalidate {
			// TODO 2 invalidation request during the same cycle
			switch state {
			case invalid:
				// Nothing
				evt.Commit()
			case modified:
				if readPending != none || writePending != none {
					// Write is ongoing, we should wait before write-back
					event.response.Notify(busResponseEvent{true})
					continue
				}

				evt.Commit()
				// Write back
				cc.snoopBusy[event.alignedAddr] = true
				cycles := latency.MemoryAccess
				cc.snoops = append(cc.snoops, co.New(func(struct{}) bool {
					if cycles > 0 {
						cycles--
						event.response.Notify(busResponseEvent{true})
						return true
					}

					memory, exists := cc.l1d.GetCacheLine(event.alignedAddr)
					if !exists {
						panic("memory address should exist")
					}
					cc.mmu.writeToMemory(event.alignedAddr, memory)

					cc.msi[event.alignedAddr] = invalid
					cc.snoopBusy[event.alignedAddr] = false
					return false
				}))
			case shared:
				if readPending != none || writePending != none {
					// Write is ongoing, we should wait before write-back
					event.response.Notify(busResponseEvent{true})
					continue
				}

				// Evict
				_, evicted := cc.l1d.EvictCacheLine(event.alignedAddr)
				if !evicted {
					panic("invalid state")
				}
				cc.msi[event.alignedAddr] = invalid
				evt.Commit()
			}
		} else {
			panic(state)
		}
	}

	return false
}

func getAlignedMemoryAddress(addrs []int32) int32 {
	addr := addrs[0]
	// TODO Divide?
	return addr - (addr % l1DCacheLineSize)
}

func (cc *cacheController) addPendingIntention(alignedAddr int32, action actionType, intention pendingState) {
	if alignedAddr%l1DCacheLineSize != 0 {
		panic(alignedAddr)
	}

	v, exists := cc.pending[alignedAddr]
	if !exists {
		v = make(map[actionType]pendingState)
		cc.pending[alignedAddr] = v
	}
	v[action] = intention
}

func (cc *cacheController) deletePendingIntention(alignedAddr int32, action actionType) {
	if alignedAddr%l1DCacheLineSize != 0 {
		panic(alignedAddr)
	}

	delete(cc.pending[alignedAddr], action)
}

func (cc *cacheController) getPendingIntention(alignedAddr int32) (read, write pendingState) {
	if alignedAddr%l1DCacheLineSize != 0 {
		panic(alignedAddr)
	}

	v, exists := cc.pending[alignedAddr]
	if !exists {
		return none, none
	}

	return v[readAction], v[writeAction]
}

func (cc *cacheController) coRead(r ccReadReq) ccReadResp {
	alignedAddr := getAlignedMemoryAddress(r.addrs)
	state := cc.msi[alignedAddr]
	switch state {
	case invalid:
		cc.read.Checkpoint(cc.coMemoryRead)
		return ccReadResp{}
	case modified, shared:
		cc.addPendingIntention(alignedAddr, readAction, pendingL1Access)
		memory := cc.getFromL1(r.addrs)
		cycles := latency.L1Access
		return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
			if cycles > 0 {
				cycles--
				return ccReadResp{}
			}
			cc.deletePendingIntention(alignedAddr, readAction)
			return ccReadResp{memory, true}
		})
	default:
		panic(state)
	}
}

type busRequestEvent struct {
	id int
	// Mutually exclusive
	read        bool
	write       bool
	invalidate  bool
	alignedAddr int32
	response    *broadcast.Relay[busResponseEvent]
}

type busResponseEvent struct {
	wait bool
}

func shouldWait(listener *broadcast.Listener[busResponseEvent]) bool {
	wait := false
loop:
	for {
		select {
		case evt := <-listener.Ch():
			wait = wait || evt.wait
		default:
			break loop
		}
	}
	return wait
}

func (cc *cacheController) coMemoryRead(r ccReadReq) ccReadResp {
	response := broadcast.NewRelay[busResponseEvent]()
	alignedAddr := getAlignedMemoryAddress(r.addrs)
	fmt.Println(cc.id, "read request", alignedAddr)
	cc.bus.Notify(busRequestEvent{
		id:          cc.id,
		read:        true,
		alignedAddr: alignedAddr,
		response:    response,
	})
	listener := response.Listener(maxCacheControllers)
	cc.addPendingIntention(alignedAddr, readAction, pendingReadRequestResponse)
	cc.read.Checkpoint(func(r ccReadReq) ccReadResp {
		if shouldWait(listener) {
			return ccReadResp{}
		}

		cycles := latency.MemoryAccess
		lineAddr, data := cc.mmu.fetchCacheLine(r.addrs[0], l1DCacheLineSize)
		cc.addPendingIntention(alignedAddr, readAction, pendingFetchCacheLineFromMemory)
		return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
			if cycles > 0 {
				cycles--
				return ccReadResp{}
			}
			cc.pushLineToL1(lineAddr, data)

			data = cc.getFromL1(r.addrs)
			cycles = latency.L1Access
			cc.addPendingIntention(alignedAddr, readAction, pendingFetchCacheLineFromMemory)
			return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
				if cycles > 0 {
					cycles--
					return ccReadResp{}
				}

				cc.read.Reset()
				cc.msi[alignedAddr] = shared
				cc.deletePendingIntention(alignedAddr, readAction)
				return ccReadResp{
					data: data,
					done: true,
				}
			})
		})
	})
	return ccReadResp{}
}

func (cc *cacheController) coWrite(r ccWriteReq) ccWriteResp {
	alignedAddr := getAlignedMemoryAddress(r.addrs)
	state := cc.msi[alignedAddr]
	switch state {
	case invalid:
		return cc.coWriteRequest(r)
	case modified:
		cycles := latency.L1Access
		cc.addPendingIntention(alignedAddr, writeAction, pendingL1Access)
		return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
			if cycles > 0 {
				cycles--
				return ccWriteResp{}
			}

			if cc.msi[alignedAddr] != modified {
				panic("invalid state")
			}
			cc.writeToL1(r.addrs, r.data)
			cc.write.Reset()
			return ccWriteResp{done: true}
		})
	case shared:
		return cc.coInvalidationRequest(r)
	default:
		panic(state)
	}
}

func (cc *cacheController) coWriteRequest(r ccWriteReq) ccWriteResp {
	alignedAddr := getAlignedMemoryAddress(r.addrs)
	fmt.Println(cc.id, "write request", alignedAddr)
	response := broadcast.NewRelay[busResponseEvent]()
	cc.bus.Notify(busRequestEvent{
		id:          cc.id,
		write:       true,
		alignedAddr: alignedAddr,
		response:    response,
	})
	listener := response.Listener(maxCacheControllers)
	cc.addPendingIntention(alignedAddr, writeAction, pendingWriteRequestResponse)
	cc.write.Checkpoint(func(r ccWriteReq) ccWriteResp {
		if shouldWait(listener) {
			return ccWriteResp{}
		}

		cycles := latency.MemoryAccess
		addr, line := cc.mmu.fetchCacheLine(r.addrs[0], l1DCacheLineSize)
		cc.addPendingIntention(alignedAddr, writeAction, pendingFetchCacheLineFromMemory)
		return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
			if cycles > 0 {
				cycles--
				return ccWriteResp{}
			}

			cycles = latency.L1Access
			evicted := cc.l1d.PushLine(addr, line)
			if len(evicted) != 0 {
				panic("not handled")
			}
			cc.addPendingIntention(alignedAddr, writeAction, pendingL1Access)
			return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
				if cycles > 0 {
					cycles--
					return ccWriteResp{}
				}

				cycles = latency.L1Access
				cc.addPendingIntention(alignedAddr, writeAction, pendingL1Access)
				return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
					if cycles > 0 {
						cycles--
						return ccWriteResp{}
					}

					cc.writeToL1(r.addrs, r.data)
					cc.msi[alignedAddr] = modified
					cc.write.Reset()
					cc.deletePendingIntention(alignedAddr, writeAction)
					return ccWriteResp{done: true}
				})
			})
		})
	})
	return ccWriteResp{}
}

func (cc *cacheController) coInvalidationRequest(r ccWriteReq) ccWriteResp {
	alignedAddr := getAlignedMemoryAddress(r.addrs)
	fmt.Println(cc.id, "invalidation request", alignedAddr)
	response := broadcast.NewRelay[busResponseEvent]()
	cc.bus.Notify(busRequestEvent{
		id:          cc.id,
		invalidate:  true,
		alignedAddr: alignedAddr,
		response:    response,
	})
	listener := response.Listener(maxCacheControllers)
	cc.addPendingIntention(alignedAddr, writeAction, pendingInvalidationRequestResponse)
	cc.write.Checkpoint(func(r ccWriteReq) ccWriteResp {
		if shouldWait(listener) {
			return ccWriteResp{}
		}

		cycles := latency.L1Access
		cc.addPendingIntention(alignedAddr, writeAction, pendingL1Access)
		return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
			if cycles > 0 {
				cycles--
				return ccWriteResp{}
			}

			cc.writeToL1(r.addrs, r.data)
			cc.msi[alignedAddr] = modified
			cc.write.Reset()
			cc.deletePendingIntention(alignedAddr, writeAction)
			return ccWriteResp{done: true}
		})
	})
	return ccWriteResp{}
}

func (cc *cacheController) pushLineToL1(addr int32, line []int8) {
	evicted := cc.l1d.PushLine(addr, line)
	if len(evicted) == 0 {
		return
	}

	panic("not handled")
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
	delta++
	fmt.Println(delta)
	fmt.Printf("write, id=%d, addr=%d, data=%d\n", cc.id, addrs[0], risc.I32FromBytes(data[0], data[1], data[2], data[3]))
	cc.l1d.Write(addrs[0], data)
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
	// TODO Really?
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

func (cc *cacheController) isEmpty() bool {
	return cc.read.IsStart() && cc.write.IsStart() && cc.snoop.IsStart()
}
