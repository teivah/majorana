package mvp6_1

import (
	"sort"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type memoryManagementUnit struct {
	ctx *risc.Context
	l1i *comp.LRUCache
	l1d *comp.LRUCache
}

func newMemoryManagementUnit(ctx *risc.Context) *memoryManagementUnit {
	return &memoryManagementUnit{
		ctx: ctx,
		l1i: comp.NewLRUCache(l1ICacheLineSize, liICacheSize),
		l1d: comp.NewLRUCache(l1DCacheLineSize, liDCacheSize),
	}
}

func (u *memoryManagementUnit) getFromL1I(addrs []int32) ([]int8, bool) {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		v, exists := u.l1i.Get(addr)
		if !exists {
			return nil, false
		}
		memory = append(memory, v)
	}
	return memory, true
}

func (u *memoryManagementUnit) pushLineToL1I(addr int32, line []int8) {
	u.l1i.PushLine(addr, line)
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

func (u *memoryManagementUnit) doesExecutionMemoryChangesExistsInL1D(execution risc.Execution) bool {
	addrs := make([]int32, 0, len(execution.MemoryChanges))
	for addr := range execution.MemoryChanges {
		addrs = append(addrs, addr)
	}
	_, exists := u.getFromL1D(addrs)
	return exists
}

func (u *memoryManagementUnit) writeExecutionMemoryChangesToL1D(execution risc.Execution) {
	type change struct {
		addr   int32
		change int8
	}
	var changes []change
	for a, v := range execution.MemoryChanges {
		changes = append(changes, change{
			addr:   a,
			change: v,
		})
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].addr < changes[j].addr
	})
	data := make([]int8, 0, len(changes))
	for _, c := range changes {
		data = append(data, c.change)
	}
	u.writeToL1D(changes[0].addr, data)
}

func (u *memoryManagementUnit) getFromMemory(addrs []int32) []int8 {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		memory = append(memory, u.ctx.Memory[addr])
	}
	return memory
}

func (u *memoryManagementUnit) fetchCacheLine(addr int32) []int8 {
	memory := make([]int8, 0, l1DCacheLineSize)
	for i := 0; i < l1DCacheLineSize; i++ {
		if int(addr)+i >= len(u.ctx.Memory) {
			memory = append(memory, 0)
		} else {
			memory = append(memory, u.ctx.Memory[int(addr)+i])
		}
	}
	return memory
}

func (u *memoryManagementUnit) pushLineToL1D(addr int32, line []int8) {
	evicted := u.l1d.PushLine(addr, line)
	if len(evicted) == 0 {
		return
	}
	u.writeToMemory(addr, line)
}

func (u *memoryManagementUnit) writeToL1D(addr int32, data []int8) {
	u.l1d.Write(addr, data)
}

func (u *memoryManagementUnit) writeToMemory(addr int32, data []int8) {
	for i, v := range data {
		if int(addr)+i >= len(u.ctx.Memory) {
			return
		}
		u.ctx.Memory[addr+int32(i)] = v
	}
}

func (u *memoryManagementUnit) flush() int {
	additionalCycles := 0
	for _, line := range u.l1d.Lines() {
		additionalCycles += cyclesMemoryAccess
		for i := 0; i < l1DCacheLineSize; i++ {
			u.writeToMemory(line.Boundary[0], line.Data)
		}
	}
	return additionalCycles
}
