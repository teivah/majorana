package mvp7_0

import (
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type ccReadReq struct {
	cycle int
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
	read      co.Coroutine[ccReadReq, ccReadResp]
	write     co.Coroutine[ccWriteReq, ccWriteResp]
	snoop     co.Coroutine[struct{}, struct{}]
	msi       *msi
	rlockSems map[comp.AlignedAddress]*comp.Sem
	lockSems  map[comp.AlignedAddress]*comp.Sem
	// Transient
	post func()
}

func newCacheController(id int, ctx *risc.Context, mmu *memoryManagementUnit, msi *msi) *cacheController {
	cc := &cacheController{
		ctx:       ctx,
		id:        id,
		mmu:       mmu,
		l1d:       comp.NewLRUCache(l1DCacheLineSize, l1DCacheSize),
		msi:       msi,
		rlockSems: make(map[comp.AlignedAddress]*comp.Sem),
		lockSems:  make(map[comp.AlignedAddress]*comp.Sem),
	}
	cc.read = co.New(cc.coRead)
	cc.write = co.New(cc.coWrite)
	cc.snoop = co.New(cc.coSnoop)
	return cc
}

// coSnoop is the coroutine executed *before* coRead and coWrite to execute
// the requests sent by msi.
func (cc *cacheController) coSnoop(struct{}) struct{} {
	requests := cc.msi.getPendingRequestsToCore(cc.id)
	if len(requests) == 0 {
		return struct{}{}
	}

	for req, info := range requests {
		switch req.request {
		case evict:
			cc.snoop.Append(func(struct{}) bool {
				_, _ = cc.l1d.EvictCacheLine(req.alignedAddr)
				info.done()
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
				info.done()
				return true
			})
		default:
			panic(req.request)
		}
	}
	return struct{}{}
}

func getAlignedMemoryAddress(addrs []int32) comp.AlignedAddress {
	addr := addrs[0]
	return comp.AlignedAddress(addr - (addr % l1DCacheLineSize))
}

func (cc *cacheController) coRead(r ccReadReq) ccReadResp {
	resp, post, sem := cc.msi.rLock(cc.id, r.addrs)
	if resp.wait {
		return ccReadResp{}
	}
	cc.rlockSems[getAlignedMemoryAddress(r.addrs)] = sem
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
					delete(cc.rlockSems, getAlignedMemoryAddress(r.addrs))
					return ccReadResp{memory, true}
				})
			} else if resp.fetchFromMemory {
				if _, exists := cc.l1d.GetCacheLine(getAlignedMemoryAddress(r.addrs)); exists {
					panic("invalid state")
				}
				cycles := latency.MemoryAccess
				lineAddr, data := cc.mmu.fetchCacheLine(r.addrs[0], l1DCacheLineSize)
				return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
					if cycles > 0 {
						cycles--
						return ccReadResp{}
					}
					shouldEvict := cc.pushLineToL1(lineAddr, data)
					if shouldEvict != nil {
						pending := cc.msi.evictExtraCacheLine(cc.id, shouldEvict.Boundary[0])
						cc.read.Checkpoint(func(r ccReadReq) ccReadResp {
							if pending != nil && !pending.isDone() {
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
								delete(cc.rlockSems, getAlignedMemoryAddress(r.addrs))
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
						delete(cc.rlockSems, getAlignedMemoryAddress(r.addrs))
						return ccReadResp{data, true}
					})
				})
			} else {
				panic("invalid state")
			}
		})
	})
}

func (cc *cacheController) coWrite(r ccWriteReq) ccWriteResp {
	resp, post, sem := cc.msi.lock(cc.id, r.addrs)
	if resp.wait {
		return ccWriteResp{}
	}
	cc.lockSems[getAlignedMemoryAddress(r.addrs)] = sem
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
					pending := cc.msi.evictExtraCacheLine(cc.id, shouldEvict.Boundary[0])
					cycles = latency.L1Access
					cc.write.Checkpoint(func(r ccWriteReq) ccWriteResp {
						if pending != nil && !pending.isDone() {
							return ccWriteResp{}
						}
						if cycles > 0 {
							cycles--
							return ccWriteResp{}
						}
						cc.post = post
						return cc.write.ExecuteWithCheckpoint(r, cc.coWriteToL1)
					})
					return ccWriteResp{}
				}
				cc.post = post
				return cc.write.ExecuteWithCheckpoint(r, cc.coWriteToL1)
			})
		} else if resp.writeToL1 {
			cc.post = post
			return cc.write.ExecuteWithCheckpoint(r, cc.coWriteToL1)
		} else {
			panic("invalid state")
		}
	})
}

// coWriteToL1 is called only if the line is already fetched.
func (cc *cacheController) coWriteToL1(r ccWriteReq) ccWriteResp {
	cycles := latency.L1Access
	return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
		if cycles > 0 {
			cycles--
			return ccWriteResp{}
		}

		cc.writeToL1(r.addrs, r.data)
		cc.post()
		cc.post = nil
		cc.write.Reset()
		delete(cc.lockSems, getAlignedMemoryAddress(r.addrs))
		return ccWriteResp{done: true}
	})
}

func (cc *cacheController) pushLineToL1(addr comp.AlignedAddress, line []int8) *comp.Line {
	if cc.isAddressInL1([]int32{int32(addr)}) {
		// No need to wait if it was already in L1
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
	for k, sem := range cc.rlockSems {
		sem.RUnlock()
		delete(cc.rlockSems, k)
	}
	for k, sem := range cc.lockSems {
		sem.Unlock()
		delete(cc.rlockSems, k)
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
