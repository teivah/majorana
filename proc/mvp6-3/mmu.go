package mvp6_3

import (
	"sort"

	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type memoryManagementUnit struct {
	ctx      *risc.Context
	l1i      *comp.LRUCache
	l3       *comp.LRUCache
	pendings [][2]int32
}

func newMemoryManagementUnit(ctx *risc.Context) *memoryManagementUnit {
	return &memoryManagementUnit{
		ctx: ctx,
		l1i: comp.NewLRUCache(l1ICacheLineSize, l1ICacheSize),
		l3:  comp.NewLRUCache(l3CacheLineSize, l3CacheSize),
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

func (u *memoryManagementUnit) pushLineToL1I(addr comp.AlignedAddress, line []int8) {
	u.l1i.PushLine(addr, line)
}

// getFromL3 returns whether an address is pending request and present in L3.
func (u *memoryManagementUnit) getFromL3(addrs []int32) ([]int8, bool, bool) {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		v, exists := u.l3.Get(addr)
		if !exists {
			for _, pending := range u.pendings {
				if pending[0] <= addr && addr < pending[1] {
					return nil, true, false
				}
			}

			u.pendings = append(u.pendings, [2]int32{addr, addr + l3CacheLineSize + 1})
			return nil, false, false
		}
		memory = append(memory, v)
	}
	return memory, false, true
}

func (u *memoryManagementUnit) doesExecutionMemoryChangesExistsInL3(execution risc.Execution) bool {
	addrs := make([]int32, 0, len(execution.MemoryChanges))
	for addr := range execution.MemoryChanges {
		addrs = append(addrs, addr)
	}
	_, _, exists := u.getFromL3(addrs)
	return exists
}

func (u *memoryManagementUnit) writeExecutionMemoryChangesToL3(execution risc.Execution) {
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
	u.writeToL3(changes[0].addr, data)
}

func (u *memoryManagementUnit) getFromMemory(addrs []int32) []int8 {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		memory = append(memory, u.ctx.Memory[addr])
	}
	return memory
}

func (u *memoryManagementUnit) fetchCacheLine(addr int32) []int8 {
	memory := make([]int8, 0, l3CacheLineSize)
	for i := 0; i < l3CacheLineSize; i++ {
		if int(addr)+i >= len(u.ctx.Memory) {
			memory = append(memory, 0)
		} else {
			memory = append(memory, u.ctx.Memory[int(addr)+i])
		}
	}
	return memory
}

func (u *memoryManagementUnit) pushLineToL3(addr comp.AlignedAddress, line []int8) {
	evicted := u.l3.PushLine(addr, line)
	for i, pending := range u.pendings {
		if pending[0] == int32(addr) {
			if len(u.pendings) == 0 {
				u.pendings = nil
			} else {
				u.pendings = append(u.pendings[:i], u.pendings[i+1:]...)
			}
			break
		}
	}
	if len(evicted) == 0 {
		return
	}
	u.writeToMemory(int32(addr), line)
}

func (u *memoryManagementUnit) writeToL3(addr int32, data []int8) {
	u.l3.Write(addr, data)
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
	for _, line := range u.l3.Lines() {
		additionalCycles += latency.MemoryAccess
		for i := 0; i < l3CacheLineSize; i++ {
			u.writeToMemory(int32(line.Boundary[0]), line.Data)
		}
	}
	return additionalCycles
}
