package mvp6_4

import (
	"github.com/teivah/majorana/risc"
)

type memoryManagementUnit struct {
	ctx *risc.Context
}

func newMemoryManagementUnit(ctx *risc.Context, eu int) *memoryManagementUnit {
	return &memoryManagementUnit{
		ctx: ctx,
	}
}

func (u *memoryManagementUnit) getFromMemory(addrs []int32) []int8 {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		memory = append(memory, u.ctx.Memory[addr])
	}
	return memory
}

func (u *memoryManagementUnit) fetchCacheLine(addr int32, cacheLineSize int32) (int32, []int8) {
	// Starting address must be a multiple of the cache line length
	var cacheLineAddr int32
	if addr%cacheLineSize == 0 {
		cacheLineAddr = addr
	} else {
		cacheLineAddr = addr - (addr % cacheLineSize)
	}

	memory := make([]int8, 0, cacheLineSize)
	for i := 0; i < int(cacheLineSize); i++ {
		if int(cacheLineAddr)+i >= len(u.ctx.Memory) {
			memory = append(memory, 0)
		} else {
			memory = append(memory, u.ctx.Memory[int(cacheLineAddr)+i])
		}
	}
	return cacheLineAddr, memory
}

func (u *memoryManagementUnit) writeToMemory(addr int32, data []int8) {
	for i, v := range data {
		if int(addr)+i >= len(u.ctx.Memory) {
			return
		}
		u.ctx.Memory[addr+int32(i)] = v
	}
}
