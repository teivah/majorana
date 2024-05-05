package mvp7_2

import (
	"fmt"

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
	ctx         *risc.Context
	id          int
	mmu         *memoryManagementUnit
	l1d         *comp.LRUCache
	l3          *comp.LRUCache
	read        co.Coroutine[ccReadReq, ccReadResp]
	write       co.Coroutine[ccWriteReq, ccWriteResp]
	snoop       co.Coroutine[struct{}, struct{}]
	msi         *msi
	l1RLockSems map[comp.AlignedAddress]*comp.Sem
	l1LockSems  map[comp.AlignedAddress]*comp.Sem

	// Transient
	post func()
}

func newCacheController(id int, ctx *risc.Context, mmu *memoryManagementUnit, msi *msi, l3 *comp.LRUCache) *cacheController {
	cc := &cacheController{
		ctx:         ctx,
		id:          id,
		mmu:         mmu,
		l1d:         comp.NewLRUCache(l1DCacheLineSize, l1DCacheSize),
		l3:          l3,
		msi:         msi,
		l1RLockSems: make(map[comp.AlignedAddress]*comp.Sem),
		l1LockSems:  make(map[comp.AlignedAddress]*comp.Sem),
	}
	cc.read = co.New(cc.coRead)
	cc.write = co.New(cc.coWrite)
	cc.snoop = co.New(cc.coSnoop)
	return cc
}

func (cc *cacheController) assertAddrInState(addr comp.AlignedAddress, expected msiState) {
	got := cc.msi.states[msiEntry{
		id:          cc.id,
		alignedAddr: addr,
	}]
	if expected != got {
		panic(fmt.Sprintf("invalid state: expected %s, got %s", expected, got))
	}
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
		case l1Evict:
			cc.assertAddrInState(req.alignedAddr, shared)
			cc.msi.staleState = true
			cc.snoop.Append(func(struct{}) bool {
				_, _ = cc.l1d.EvictCacheLine(req.alignedAddr)
				info.done()
				return true
			})
		case l3Evict:
			cc.msi.staleState = true
			cc.snoop.Append(func(struct{}) bool {
				mu := cc.msi.getL3Lock([]int32{int32(req.alignedAddr)})
				if mu.TryLock() {
					return false
				}

				_, _ = cc.l3.EvictCacheLine(req.alignedAddr)
				cc.msi.l3ReleaseWriteNotify(req.alignedAddr)
				info.done()
				mu.Unlock()
				return true
			})
		case l1WriteBack:
			cc.assertAddrInState(req.alignedAddr, modified)
			cycles1 := latency.L3Access
			cycles2 := latency.MemoryAccess
			cycles3 := latency.L3Access
			cc.snoop.Append(func(struct{}) bool {
				if cycles1 > 0 {
					cycles1--
					return false
				}

				memory, exists := cc.l1d.GetCacheLine(req.alignedAddr)
				if !exists {
					panic("memory address should exist")
				}

				if !cc.isAddressInL3([]int32{int32(req.alignedAddr)}) {
					// Cache line was evicted
					if cycles2 > 0 {
						cycles2--
						return false
					}
					cc.mmu.writeToMemory(req.alignedAddr, memory)
					_, evicted := cc.l1d.EvictCacheLine(req.alignedAddr)
					if !evicted {
						panic("invalid state")
					}
					info.done()
					return true
				} else {
					if cycles3 > 0 {
						cycles3--
						return false
					}
					cc.writeToL3(req.alignedAddr, memory)
					_, evicted := cc.l1d.EvictCacheLine(req.alignedAddr)
					if !evicted {
						panic("invalid state")
					}
					info.done()
					return true
				}
			})
		case l3WriteBack:
			cycles := latency.MemoryAccess
			cc.snoop.Append(func(struct{}) bool {
				if cycles > 0 {
					cycles--
					return false
				}

				mu := cc.msi.getL3Lock([]int32{int32(req.alignedAddr)})
				if mu.TryLock() {
					return false
				}

				memory, exists := cc.l3.GetCacheLine(req.alignedAddr)
				if !exists {
					panic("memory address should exist")
				}

				cc.mmu.writeToMemory(req.alignedAddr, memory)
				_, evicted := cc.l3.EvictCacheLine(req.alignedAddr)
				cc.msi.l3ReleaseWriteNotify(req.alignedAddr)
				if !evicted {
					panic("invalid state")
				}
				info.done()
				mu.Unlock()
				return true
			})
		default:
			panic(req.request)
		}
	}
	return struct{}{}
}

func getL1AlignedMemoryAddress(addrs []int32) comp.AlignedAddress {
	return getAlignedMemoryAddress(addrs, l1DCacheLineSize)
}

func getL3AlignedMemoryAddress(addrs []int32) comp.AlignedAddress {
	return getAlignedMemoryAddress(addrs, l3CacheLineSize)
}

func getAlignedMemoryAddress(addrs []int32, align int32) comp.AlignedAddress {
	addr := addrs[0]
	return comp.AlignedAddress(addr - (addr % align))
}

func (cc *cacheController) coRead(r ccReadReq) ccReadResp {
	resp, post, sem := cc.msi.rLock(cc.id, r.addrs)
	if resp.wait {
		return ccReadResp{}
	}
	cc.post = post
	cc.l1RLockSems[getL1AlignedMemoryAddress(r.addrs)] = sem
	return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
		for _, pending := range resp.pendings {
			if !pending.isDone() {
				return ccReadResp{}
			}
		}

		return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
			if resp.fromL1 {
				return cc.read.ExecuteWithCheckpoint(r, cc.coReadFromL1)
			} else if resp.notFromL1 {
				if _, exists := cc.l1d.GetCacheLine(getL1AlignedMemoryAddress(r.addrs)); exists {
					panic("invalid state")
				}

				cycles := latency.L3Access
				return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
					if cycles > 0 {
						cycles--
						return ccReadResp{}
					}

					if cc.isAddressInL3(r.addrs) {
						// Fetch from L3, sync to L1
						l1Addr, l1Data, exists := cc.l3.GetSubCacheLine(r.addrs, l1DCacheLineSize)
						if !exists {
							panic("invalid state")
						}

						shouldEvict := cc.pushLineToL1(l1Addr, l1Data)
						if shouldEvict != nil {
							pending := cc.msi.evictL1ExtraCacheLine(cc.id, shouldEvict.Boundary[0])
							cc.read.Checkpoint(func(r ccReadReq) ccReadResp {
								if pending != nil && !pending.isDone() {
									return ccReadResp{}
								}
								return cc.read.ExecuteWithCheckpoint(r, cc.coReadFromL1)
							})
							return ccReadResp{}
						}
						return cc.read.ExecuteWithCheckpoint(r, cc.coReadFromL1)
					} else {
						// Fetch from memory, sync to L3, sync to L1
						l3Addr, l3Data := cc.mmu.fetchCacheLine(r.addrs[0], l3CacheLineSize)
						cycles := latency.MemoryAccess
						return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
							if cycles > 0 {
								cycles--
								return ccReadResp{}
							}

							return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
								mu := cc.msi.getL3Lock(r.addrs)
								if !mu.TryLock() {
									return ccReadResp{}
								}

								cycles = latency.L3Access
								return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
									if cycles > 0 {
										cycles--
										return ccReadResp{}
									}
									shouldEvict := cc.pushLineToL3(l3Addr, l3Data)
									mu.Unlock()
									if shouldEvict != nil {
										pending := cc.msi.evictL3ExtraCacheLine(cc.id, shouldEvict.Boundary[0])
										cc.read.Checkpoint(func(r ccReadReq) ccReadResp {
											if pending != nil && !pending.isDone() {
												return ccReadResp{}
											}
											return cc.read.ExecuteWithCheckpoint(r, cc.coSyncReadFromL1)
										})
									}
									return cc.read.ExecuteWithCheckpoint(r, cc.coSyncReadFromL1)
								})
							})
						})
					}
				})
			} else {
				panic("invalid state")
			}
		})
	})
}

func (cc *cacheController) coSyncReadFromL1(r ccReadReq) ccReadResp {
	l1Addr, l1Data, exists := cc.l3.GetSubCacheLine(r.addrs, l1DCacheLineSize)
	if !exists {
		panic("invalid state")
	}

	shouldEvict := cc.pushLineToL1(l1Addr, l1Data)
	if shouldEvict != nil {
		pending := cc.msi.evictL1ExtraCacheLine(cc.id, shouldEvict.Boundary[0])
		cc.read.Checkpoint(func(r ccReadReq) ccReadResp {
			if pending != nil && !pending.isDone() {
				return ccReadResp{}
			}

			return cc.read.ExecuteWithCheckpoint(r, cc.coReadFromL1)
		})
		return ccReadResp{}
	}
	return cc.read.ExecuteWithCheckpoint(r, cc.coReadFromL1)
}

func (cc *cacheController) coReadFromL1(r ccReadReq) ccReadResp {
	data := cc.getFromL1(r.addrs)
	cycles := latency.L1Access
	return cc.read.ExecuteWithCheckpoint(r, func(r ccReadReq) ccReadResp {
		if cycles > 0 {
			cycles--
			return ccReadResp{}
		}
		cc.post()
		cc.post = nil
		cc.read.Reset()
		delete(cc.l1RLockSems, getL1AlignedMemoryAddress(r.addrs))
		return ccReadResp{data, true}
	})
}

func (cc *cacheController) coWrite(r ccWriteReq) ccWriteResp {
	resp, post, sem := cc.msi.lock(cc.id, r.addrs)
	if resp.wait {
		return ccWriteResp{}
	}
	cc.post = post
	cc.l1LockSems[getL1AlignedMemoryAddress(r.addrs)] = sem
	return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
		for _, pending := range resp.pendings {
			if !pending.isDone() {
				return ccWriteResp{}
			}
		}

		if resp.notFromL1 {
			if cc.isAddressInL3(r.addrs) {
				// Fetch from L3, sync to L1
				l1Addr, l1Data, exists := cc.l3.GetSubCacheLine(r.addrs, l1DCacheLineSize)
				if !exists {
					panic("invalid state")
				}

				cycles := latency.L1Access
				return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
					if cycles > 0 {
						cycles--
						return ccWriteResp{}
					}

					shouldEvict := cc.pushLineToL1(l1Addr, l1Data)
					if shouldEvict != nil {
						pending := cc.msi.evictL1ExtraCacheLine(cc.id, shouldEvict.Boundary[0])
						cycles := latency.L1Access
						cc.write.Checkpoint(func(r ccWriteReq) ccWriteResp {
							if pending != nil && !pending.isDone() {
								return ccWriteResp{}
							}
							if cycles > 0 {
								cycles--
								return ccWriteResp{}
							}
							return cc.write.ExecuteWithCheckpoint(r, cc.coWriteToL1)
						})
						return ccWriteResp{}
					}
					return cc.write.ExecuteWithCheckpoint(r, cc.coWriteToL1)
				})
			} else {
				// Fetch from memory, sync to L3, sync to L1
				cycles := latency.MemoryAccess
				l3Addr, l3Data := cc.mmu.fetchCacheLine(r.addrs[0], l3CacheLineSize)
				return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
					if cycles > 0 {
						cycles--
						return ccWriteResp{}
					}

					cycles = latency.L3Access
					return cc.write.ExecuteWithCheckpoint(r, func(r ccWriteReq) ccWriteResp {
						if cycles > 0 {
							cycles--
							return ccWriteResp{}
						}
						mu := cc.msi.getL3Lock(r.addrs)
						if !mu.TryLock() {
							return ccWriteResp{}
						}

						mu.Unlock()
						shouldEvict := cc.pushLineToL3(l3Addr, l3Data)
						if shouldEvict != nil {
							pending := cc.msi.evictL3ExtraCacheLine(cc.id, shouldEvict.Boundary[0])
							cc.write.Checkpoint(func(r ccWriteReq) ccWriteResp {
								if pending != nil && !pending.isDone() {
									return ccWriteResp{}
								}
								return cc.write.ExecuteWithCheckpoint(r, cc.coSyncWriteToL1)
							})
							return ccWriteResp{}
						}
						return cc.write.ExecuteWithCheckpoint(r, cc.coSyncWriteToL1)
					})
				})
			}
		} else if resp.writeToL1 {
			return cc.write.ExecuteWithCheckpoint(r, cc.coWriteToL1)
		} else {
			panic("invalid state")
		}
	})
}

func (cc *cacheController) coSyncWriteToL1(r ccWriteReq) ccWriteResp {
	l1Addr, l1Data, exists := cc.l3.GetSubCacheLine(r.addrs, l1DCacheLineSize)
	if !exists {
		panic("invalid state")
	}

	shouldEvict := cc.pushLineToL1(l1Addr, l1Data)
	if shouldEvict != nil {
		pending := cc.msi.evictL1ExtraCacheLine(cc.id, shouldEvict.Boundary[0])
		cycles := latency.L1Access
		cc.write.Checkpoint(func(r ccWriteReq) ccWriteResp {
			if pending != nil && !pending.isDone() {
				return ccWriteResp{}
			}
			if cycles > 0 {
				cycles--
				return ccWriteResp{}
			}
			return cc.write.ExecuteWithCheckpoint(r, cc.coWriteToL1)
		})
		return ccWriteResp{}
	}
	return cc.write.ExecuteWithCheckpoint(r, cc.coWriteToL1)
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
		delete(cc.l1LockSems, getL1AlignedMemoryAddress(r.addrs))
		return ccWriteResp{done: true}
	})
}

func (cc *cacheController) pushLineToL1(addr comp.AlignedAddress, line []int8) *comp.Line {
	if len(line) != l1DCacheLineSize || addr%l1DCacheLineSize != 0 {
		panic("invalid state")
	}
	if cc.isAddressInL1([]int32{int32(addr)}) {
		// No need to wait if it was already in L1
		return nil
	}
	return cc.l1d.PushLineWithEvictionWarning(addr, line)
}

func (cc *cacheController) pushLineToL3(addr comp.AlignedAddress, line []int8) *comp.Line {
	if len(line) != l3CacheLineSize || addr%l3CacheLineSize != 0 {
		panic("invalid state")
	}
	if cc.isAddressInL3([]int32{int32(addr)}) {
		// No need to wait if it was already in L3
		return nil
	}
	return cc.l3.PushLineWithEvictionWarning(addr, line)
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

func (cc *cacheController) isAddressInL3(addrs []int32) bool {
	_, exists := cc.l3.Get(addrs[0])
	return exists
}

func (cc *cacheController) getFromL3(addrs []int32) []int8 {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		v, exists := cc.l3.Get(addr)
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

func (cc *cacheController) writeToL3(l1Addr comp.AlignedAddress, data []int8) {
	l3Addr := getL3AlignedMemoryAddress([]int32{int32(l1Addr)})
	cc.msi.l3WriteNotify(l3Addr)
	cc.l3.Write(int32(l1Addr), data)
}

func (cc *cacheController) flush() {
	cc.read.Reset()
	cc.write.Reset()
	for k, sem := range cc.l1RLockSems {
		sem.RUnlock()
		delete(cc.l1RLockSems, k)
	}
	for k, sem := range cc.l1LockSems {
		sem.Unlock()
		delete(cc.l1RLockSems, k)
	}
}

func (cc *cacheController) export() int {
	additionalCycles := 0
	for _, line := range cc.l1d.ExistingLines() {
		addr := line.Boundary[0]
		if cc.msi.states[msiEntry{cc.id, addr}] != modified {
			// If not modified, we don't write back the line in memory
			continue
		}

		if cc.isAddressInL3([]int32{int32(line.Boundary[0])}) {
			mu := cc.msi.getL3Lock([]int32{int32(line.Boundary[0])})
			if !mu.TryLock() {
				panic("invalid state")
			}

			additionalCycles += latency.L3Access
			cc.writeToL3(line.Boundary[0], line.Data)
			mu.Unlock()
		} else {
			// Line was evicted
			additionalCycles += latency.MemoryAccess
			cc.mmu.writeToMemory(line.Boundary[0], line.Data)
		}
	}
	return additionalCycles
}

func (cc *cacheController) isEmpty() bool {
	return cc.read.IsStart() && cc.write.IsStart() && cc.snoop.IsStart()
}
