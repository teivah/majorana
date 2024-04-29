package mvp7_0

import (
	"github.com/teivah/majorana/risc"
)

type memoryManagementUnit struct {
	ctx *risc.Context
}

func newMemoryManagementUnit(ctx *risc.Context) *memoryManagementUnit {
	return &memoryManagementUnit{
		ctx: ctx,
	}
}

func (u *memoryManagementUnit) fetchCacheLine(addr int32, cacheLineSize int32) (int32, []int8) {
	cacheLineAddr := getAlignedMemoryAddress([]int32{addr})
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
