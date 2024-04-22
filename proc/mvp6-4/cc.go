package mvp6_4

import (
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/latency"
)

type msiState = int32

const (
	invalid msiState = iota // Default has to be invalid
	shared
	modified
)

type cacheController struct {
	mmu *memoryManagementUnit
	// msi implements the MSI protocol.
	// Source: https://www.youtube.com/watch?v=gAUVAel-2Fg.
	msi map[int32]msiState
	co.Coroutine[ccReq, ccResp]
}

func newCacheController(mmu *memoryManagementUnit) *cacheController {
	cc := &cacheController{
		mmu: mmu,
		msi: make(map[int32]msiState),
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

func (cc *cacheController) start(r ccReq) ccResp {
	memory, pending, exists := cc.mmu.getFromL3(r.addrs)
	if pending {
		return ccResp{}
	} else if exists {
		remainingCycles := latency.L3Access - 1
		cc.Checkpoint(func(r ccReq) ccResp {
			if remainingCycles > 0 {
				remainingCycles--
				return ccResp{}
			}
			cc.Reset()
			return ccResp{
				memory: memory,
				done:   true,
			}
		})
		return ccResp{}
	}

	// Read from memory
	remainingCycles := latency.MemoryAccess - 1
	cc.Checkpoint(func(r ccReq) ccResp {
		if remainingCycles > 0 {
			remainingCycles--
			return ccResp{}
		}
		cc.Reset()
		line := cc.mmu.fetchCacheLine(r.addrs[0])
		cc.mmu.pushLineToL3(r.addrs[0], line)
		m, _, exists := cc.mmu.getFromL3(r.addrs)
		if !exists {
			panic("cache line doesn't exist")
		}
		return ccResp{
			memory: m,
			done:   true,
		}
	})
	return ccResp{}
}
