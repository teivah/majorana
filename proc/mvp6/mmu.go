package mvp6

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type memoryManagementUnit struct {
	ctx *risc.Context
	l1d *comp.LRUCache
}

func newMemoryManagementUnit(ctx *risc.Context) *memoryManagementUnit {
	return &memoryManagementUnit{
		ctx: ctx,
		l1d: comp.NewLRUCache(l1DCacheLineSizeInBytes, liDCacheSizeInBytes),
	}
}

func (u *memoryManagementUnit) getFromL1D(addrs []int32) ([]int8, bool) {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		v, exists := u.l1d.Get(addr)
		if !exists {
			return nil, false
		}
		memory = append(memory, v)
	}
	return memory, true
}

func (u *memoryManagementUnit) getFromMemory(addrs []int32) []int8 {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		memory = append(memory, u.ctx.Memory[addr])
	}
	return memory
}

func (u *memoryManagementUnit) pushToL1D(addr int32, data []int8) {
	u.l1d.Push(addr, data)
}
