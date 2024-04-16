package comp

import (
	"github.com/teivah/majorana/risc"
)

// RAT is the register allocation table
type RAT struct {
	length int
	toVal  map[risc.RegisterType]int
	// TODO Slice
	toReg map[int]risc.RegisterType
	data  []int32
	next  int
}

func NewRAT(length int) *RAT {
	return &RAT{
		length: length,
		toVal:  make(map[risc.RegisterType]int),
		toReg:  make(map[int]risc.RegisterType),
		data:   make([]int32, length),
	}
}

func (r *RAT) Read(register risc.RegisterType) (int32, bool) {
	v, exists := r.toVal[register]
	if !exists {
		return 0, false
	}
	return r.data[v], true
}

func (r *RAT) Write(register risc.RegisterType, value int32) {
	reg, exists := r.toReg[r.next]

	if exists {
		if r.toVal[reg] == r.next {
			delete(r.toVal, reg)
		}
	}

	r.toVal[register] = r.next
	r.toReg[r.next] = register
	r.data[r.next] = value

	r.next++
	if r.next == r.length {
		r.next = 0
	}
}
