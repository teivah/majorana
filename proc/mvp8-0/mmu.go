package mvp8_0

import (
	"github.com/teivah/majorana/proc/comp"
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

func (u *memoryManagementUnit) fetchCacheLine(addr int32, cacheLineSize int32) (comp.AlignedAddress, []int8) {
	alignedAddr := getAlignedMemoryAddress([]int32{addr}, cacheLineSize)
	memory := make([]int8, 0, cacheLineSize)
	for i := 0; i < int(cacheLineSize); i++ {
		if int(alignedAddr)+i >= len(u.ctx.Memory) {
			memory = append(memory, 0)
		} else {
			memory = append(memory, u.ctx.Memory[int(alignedAddr)+i])
		}
	}
	return alignedAddr, memory
}

func (u *memoryManagementUnit) writeToMemory(addr comp.AlignedAddress, data []int8) {
	for i, v := range data {
		if int(addr)+i >= len(u.ctx.Memory) {
			return
		}
		u.ctx.Memory[int32(addr)+int32(i)] = v
	}
}
