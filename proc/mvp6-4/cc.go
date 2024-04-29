package mvp6_4

import (
	"github.com/teivah/broadcast"
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
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
	ctx      *risc.Context
	id       int
	mmu      *memoryManagementUnit
	l1d      *comp.LRUCache
	bus      *comp.Broadcast[busRequestEvent]
	read     co.Coroutine[ccReadReq, ccReadResp]
	write    co.Coroutine[ccWriteReq, ccWriteResp]
	snoop    co.Coroutine[struct{}, struct{}]
	msi      *msi
	rlockSem map[int32]*comp.Sem
	lockSem  map[int32]*comp.Sem
}

func newCacheController(id int, ctx *risc.Context, mmu *memoryManagementUnit, bus *comp.Broadcast[busRequestEvent], msi *msi) *cacheController {
	cc := &cacheController{
		ctx:      ctx,
		id:       id,
		mmu:      mmu,
		l1d:      comp.NewLRUCache(l1DCacheLineSize, l1DCacheSize),
		bus:      bus,
		msi:      msi,
		rlockSem: make(map[int32]*comp.Sem),
		lockSem:  make(map[int32]*comp.Sem),
	}
	cc.read = co.New(cc.coRead)
	cc.write = co.New(cc.coWrite)
	cc.snoop = co.New(cc.coSnoop)
	return cc
}

func (cc *cacheController) coSnoop(struct{}) struct{} {
	requests := cc.msi.requests(cc.id)
	if len(requests) == 0 {
		return struct{}{}
	}

	for _, req := range requests {
		switch req.request {
		case evict:
			cc.snoop.Append(func(struct{}) bool {
				_, evicted := cc.l1d.EvictCacheLine(req.alignedAddr)
				if !evicted {
					panic("invalid state")
				}
				req.done()
				//fmt.Println(cc.id, "evict", req.alignedAddr)
				return true
			})
		case writeBack:
			cycles := latency.MemoryAccess
			cc.snoop.Append(func(struct{}) bool {
				if cycles > 0 {
					cycles--
					return false
				}

				memory, exists := cc.l1d.GetCacheLine(req.alignedAddr)
				if !exists {
					panic("memory address should exist")
				}
				cc.mmu.writeToMemory(req.alignedAddr, memory)
				_, evicted := cc.l1d.EvictCacheLine(req.alignedAddr)
				if !evicted {
					panic("invalid state")
				}
				req.done()
				//fmt.Println(cc.id, "write-back", req.alignedAddr)
				return true
			})
		default:
			panic(req.request)
		}
	}
	return struct{}{}
}

func getAlignedMemoryAddress(addrs []int32) int32 {
	addr := addrs[0]
	// TODO Divide?
	return addr - (addr % l1DCacheLineSize)
}

func (cc *cacheController) coRead(r ccReadReq) ccReadResp {
	resp, post, sem := cc.msi.rLock(cc.id, r.addrs)
	if resp.wait {
		return ccReadResp{}
	}
	cc.rlockSem[getAlignedMemoryAddress(r.addrs)] = sem
	//fmt.Println("coread", cc.id, r.addrs)
	return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
		for _, pending := range resp.pendings {
			if !pending.isDone() {
				return ccReadResp{}
			}
		}

		return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
			if resp.readFromL1 {
				memory := cc.getFromL1(r.addrs)
				cycles := latency.L1Access
				return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
					if cycles > 0 {
						cycles--
						return ccReadResp{}
					}
					post()
					cc.read.Reset()
					delete(cc.rlockSem, getAlignedMemoryAddress(r.addrs))
					return ccReadResp{memory, true}
				})
			} else if resp.fetchFromMemory {
				if _, exists := cc.l1d.GetCacheLine(r.addrs[0]); exists {
					panic("invalid state")
				}
				cycles := latency.MemoryAccess
				lineAddr, data := cc.mmu.fetchCacheLine(r.addrs[0], l1DCacheLineSize)
				return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
					if cycles > 0 {
						cycles--
						return ccReadResp{}
					}
					// TODO Working???
					shouldEvict := cc.pushLineToL1(lineAddr, data)
					if shouldEvict != nil {
						pending := cc.msi.evict(cc.id, shouldEvict.Boundary[0])
						cc.read.Checkpoint(func(r ccReadReq) ccReadResp {
							if !pending.isDone() {
								return ccReadResp{}
							}
							//fmt.Println("done")

							data = cc.getFromL1(r.addrs)
							cycles = latency.L1Access
							return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
								if cycles > 0 {
									cycles--
									return ccReadResp{}
								}
								post()
								cc.read.Reset()
								delete(cc.rlockSem, getAlignedMemoryAddress(r.addrs))
								return ccReadResp{data, true}
							})
						})
						return ccReadResp{}
					}

					data = cc.getFromL1(r.addrs)
					cycles = latency.L1Access
					return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
						if cycles > 0 {
							cycles--
							return ccReadResp{}
						}
						post()
						cc.read.Reset()
						delete(cc.rlockSem, getAlignedMemoryAddress(r.addrs))
						return ccReadResp{data, true}
					})
				})
			} else {
				panic("invalid state")
			}
		})
	})
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

func (cc *cacheController) coWrite(r ccWriteReq) ccWriteResp {
	resp, post, sem := cc.msi.lock(cc.id, r.addrs)
	if resp.wait {
		return ccWriteResp{}
	}
	cc.lockSem[getAlignedMemoryAddress(r.addrs)] = sem
	//fmt.Println("cowrite", cc.id, r.addrs)
	return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
		for _, pending := range resp.pendings {
			if !pending.isDone() {
				return ccWriteResp{}
			}
		}

		if resp.fetchFromMemory {
			cycles := latency.MemoryAccess
			addr, line := cc.mmu.fetchCacheLine(r.addrs[0], l1DCacheLineSize)
			return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
				if cycles > 0 {
					cycles--
					return ccWriteResp{}
				}

				shouldEvict := cc.pushLineToL1(addr, line)
				if shouldEvict != nil {
					pending := cc.msi.evict(cc.id, shouldEvict.Boundary[0])
					cycles = latency.L1Access
					cc.write.Checkpoint(func(r ccWriteReq) ccWriteResp {
						if !pending.isDone() {
							return ccWriteResp{}
						}

						if cycles > 0 {
							cycles--
							return ccWriteResp{}
						}

						cycles = latency.L1Access
						return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
							if cycles > 0 {
								cycles--
								return ccWriteResp{}
							}

							cc.writeToL1(r.addrs, r.data)
							post()
							cc.write.Reset()
							delete(cc.lockSem, getAlignedMemoryAddress(r.addrs))
							return ccWriteResp{done: true}
						})
					})
					return ccWriteResp{}
				}

				return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
					if cycles > 0 {
						cycles--
						return ccWriteResp{}
					}

					cycles = latency.L1Access
					return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
						if cycles > 0 {
							cycles--
							return ccWriteResp{}
						}

						cc.writeToL1(r.addrs, r.data)
						post()
						cc.write.Reset()
						delete(cc.lockSem, getAlignedMemoryAddress(r.addrs))
						return ccWriteResp{done: true}
					})
				})
			})
		} else if resp.writeToL1 {
			cycles := latency.L1Access
			return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
				if cycles > 0 {
					cycles--
					return ccWriteResp{}
				}

				cc.writeToL1(r.addrs, r.data)
				post()
				cc.write.Reset()
				delete(cc.lockSem, getAlignedMemoryAddress(r.addrs))
				return ccWriteResp{done: true}
			})
		} else {
			panic("invalid state")
		}
	})
}

func (cc *cacheController) pushLineToL1(addr int32, line []int8) *comp.Line {
	if cc.isAddressInL1([]int32{addr}) {
		// TODO No need to wait if it was already in L1
		return nil
	}
	return cc.l1d.PushLineWithEvictionWarning(addr, line)
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

func (cc *cacheController) writeToL1(addrs []int32, data []int8) {
	cc.l1d.Write(addrs[0], data)
}

func (cc *cacheController) flush() {
	cc.read.Reset()
	cc.write.Reset()
	for k, sem := range cc.rlockSem {
		sem.RUnlock()
		delete(cc.rlockSem, k)
	}
	for k, sem := range cc.lockSem {
		sem.Unlock()
		delete(cc.rlockSem, k)
	}
}

func (cc *cacheController) export() int {
	additionalCycles := 0
	for _, line := range cc.l1d.Lines() {
		addr := line.Boundary[0]
		if cc.msi.states[msiEntry{cc.id, addr}] != modified {
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
