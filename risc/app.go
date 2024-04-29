package risc

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
)

type ExecutionContext struct {
	SequenceID      int32
	Execution       Execution
	InstructionType InstructionType
	WriteRegisters  []RegisterType
	ReadRegisters   []RegisterType
}

type Application struct {
	Instructions []InstructionRunner
	Labels       map[string]int32
}

type Context struct {
	Registers                   map[RegisterType]int32
	Transaction                 map[RegisterType]transactionUnit
	PendingWriteRegisters       map[RegisterType]int
	PendingReadRegisters        map[RegisterType]int
	pendingWriteMemoryIntention map[int32]map[int]struct{}
	Memory                      []int8
	Debug                       bool
	// SequenceID represents a monotonic ID for the sequence.
	// It increments during a jump.
	sequenceID     int32
	committedRAT   *comp.RAT[RegisterType, int32]
	transactionRAT *comp.RAT[RegisterType, transactionUnit]
	rat            bool
}

type transactionUnit struct {
	sequenceID int32
	value      int32
}

const ratLength = 10

func NewContext(debug bool, memoryBytes int, rat bool) *Context {
	return &Context{
		Registers:                   make(map[RegisterType]int32),
		Transaction:                 make(map[RegisterType]transactionUnit),
		PendingWriteRegisters:       make(map[RegisterType]int),
		PendingReadRegisters:        make(map[RegisterType]int),
		pendingWriteMemoryIntention: make(map[int32]map[int]struct{}),
		Memory:                      make([]int8, memoryBytes),
		Debug:                       debug,
		committedRAT:                comp.NewRAT[RegisterType, int32](ratLength),
		transactionRAT:              comp.NewRAT[RegisterType, transactionUnit](ratLength),
		rat:                         rat,
	}
}

func (ctx *Context) Flush() {
	ctx.PendingWriteRegisters = make(map[RegisterType]int)
	ctx.PendingReadRegisters = make(map[RegisterType]int)
	ctx.pendingWriteMemoryIntention = make(map[int32]map[int]struct{})
}

func (ctx *Context) SequenceID(pc int32) int32 {
	// TODO Find better
	return pc + ctx.sequenceID*1000
}

func (ctx *Context) IncSequenceID() {
	ctx.sequenceID++
}

func (ctx *Context) AddPendingWriteMemoryIntention(alignedAddr int32, id int) {
	v, exists := ctx.pendingWriteMemoryIntention[alignedAddr]
	if !exists {
		v = make(map[int]struct{})
		ctx.pendingWriteMemoryIntention[alignedAddr] = v
	}
	v[id] = struct{}{}
}

func (ctx *Context) PendingWriteMemoryIntention(alignedAddr int32) bool {
	v, exists := ctx.pendingWriteMemoryIntention[alignedAddr]
	if !exists {
		return false
	}
	return len(v) != 0
}

func (ctx *Context) DeletePendingWriteMemoryIntention(alignedAddr int32, id int) {
	v, exists := ctx.pendingWriteMemoryIntention[alignedAddr]
	if !exists {
		panic(fmt.Sprintf("do not exist %v", alignedAddr))
	}

	_, exists = v[id]
	if !exists {
		panic(fmt.Sprintf("do not exist %v %v", alignedAddr, id))
	}

	delete(v, id)
}

func (ctx *Context) WriteRegister(exe Execution) {
	ctx.Registers[exe.Register] = exe.RegisterValue
}

func (ctx *Context) TransactionWriteRegister(exe Execution, sequenceID int32) {
	ctx.Transaction[exe.Register] = transactionUnit{sequenceID, exe.RegisterValue}
}

func (ctx *Context) Commit() {
	for register, tu := range ctx.Transaction {
		ctx.Registers[register] = tu.value
	}
	ctx.Transaction = make(map[RegisterType]transactionUnit)
}

func (ctx *Context) Rollback(sequenceID int32) {
	for register, tu := range ctx.Transaction {
		if tu.sequenceID < sequenceID {
			ctx.Registers[register] = tu.value
		}
	}
	ctx.Transaction = make(map[RegisterType]transactionUnit)
}

func (ctx *Context) InitRAT() {
	for k, v := range ctx.Registers {
		ctx.committedRAT.Write(k, v)
	}
}

func (ctx *Context) TransactionRATWrite(exe Execution, sequenceID int32) {
	ctx.transactionRAT.Write(exe.Register, transactionUnit{sequenceID, exe.RegisterValue})
}

func (ctx *Context) RATCommit() {
	for register, tu := range ctx.transactionRAT.Values() {
		ctx.committedRAT.Write(register, tu.value)
	}
	ctx.transactionRAT = comp.NewRAT[RegisterType, transactionUnit](ratLength)
}

func (ctx *Context) RATRollback(sequenceID int32) {
	for register, tu := range ctx.transactionRAT.Values() {
		if tu.sequenceID < sequenceID {
			ctx.committedRAT.Write(register, tu.value)
		}
	}
	ctx.transactionRAT = comp.NewRAT[RegisterType, transactionUnit](ratLength)
}

func (ctx *Context) RATFlush() {
	for k, v := range ctx.committedRAT.Values() {
		ctx.Registers[k] = v
	}
}

func (ctx *Context) WriteMemory(exe Execution) {
	for k, v := range exe.MemoryChanges {
		ctx.Memory[k] = v
	}
}

func (ctx *Context) AddPendingRegisters(runner InstructionRunner) {
	for _, register := range runner.ReadRegisters() {
		if register == Zero {
			continue
		}
		ctx.PendingReadRegisters[register]++
	}
	for _, register := range runner.WriteRegisters() {
		if register == Zero {
			continue
		}
		ctx.PendingWriteRegisters[register]++
	}
}

func (ctx *Context) AddPendingWriteRegisters(registers []RegisterType) {
	for _, register := range registers {
		ctx.PendingWriteRegisters[register]++
	}
}

func (ctx *Context) DeletePendingRegisters(readRegisters, writeRegisters []RegisterType) {
	for _, register := range readRegisters {
		if register == Zero {
			continue
		}
		ctx.PendingReadRegisters[register]--
		if ctx.PendingReadRegisters[register] <= 0 {
			delete(ctx.PendingReadRegisters, register)
		}
	}
	for _, register := range writeRegisters {
		if register == Zero {
			continue
		}
		ctx.PendingWriteRegisters[register]--
		if ctx.PendingWriteRegisters[register] <= 0 {
			delete(ctx.PendingWriteRegisters, register)
		}
	}
}

func (ctx *Context) DeletePendingWriteRegisters(registers []RegisterType) {
	for _, register := range registers {
		ctx.PendingWriteRegisters[register]--
		if ctx.Registers[register] <= 0 {
			delete(ctx.Registers, register)
		}
	}
}

// IsWriteDataHazard returns true if there's already a register pending to be
// written among the provided list.
func (ctx *Context) IsWriteDataHazard(registers []RegisterType) bool {
	for _, register := range registers {
		if register == Zero {
			continue
		}
		if v, exists := ctx.PendingWriteRegisters[register]; exists && v > 0 {
			return true
		}
	}
	return false
}

func (ctx *Context) IsDataHazard(runner InstructionRunner) (bool, string) {
	for _, register := range runner.ReadRegisters() {
		if register == Zero {
			continue
		}
		if v, exists := ctx.PendingWriteRegisters[register]; exists && v > 0 {
			// An instruction needs to read from a register that was updated
			return true, fmt.Sprintf("Read hazard on %s", register)
		}
	}
	return false, ""
}

func (ctx *Context) IsDataHazard3(runner InstructionRunner) ([]Hazard, map[HazardType]bool) {
	var hazards []Hazard
	hazardTypes := make(map[HazardType]bool)
	for _, register := range runner.ReadRegisters() {
		if register == Zero {
			continue
		}
		if v, exists := ctx.PendingWriteRegisters[register]; exists && v > 0 {
			hazards = append(hazards, Hazard{Type: ReadAfterWrite, Register: register})
			hazardTypes[ReadAfterWrite] = true
		}
	}
	for _, register := range runner.WriteRegisters() {
		if register == Zero {
			continue
		}
		if v, exists := ctx.PendingWriteRegisters[register]; exists && v > 0 {
			hazards = append(hazards, Hazard{Type: WriteAfterWrite, Register: register})
			hazardTypes[WriteAfterWrite] = true
		}
		if v, exists := ctx.PendingReadRegisters[register]; exists && v > 0 {
			hazards = append(hazards, Hazard{Type: WriteAfterRead, Register: register})
			hazardTypes[WriteAfterRead] = true
		}
	}
	return hazards, hazardTypes
}

type HazardType uint32

const (
	// ReadAfterWrite we should wait to read the latest value
	ReadAfterWrite HazardType = iota
	// WriteAfterWrite we should get the latest written value in the end
	WriteAfterWrite
	// WriteAfterRead we should read the value before it's written
	WriteAfterRead
)

func (h HazardType) Stringer() string {
	switch h {
	case ReadAfterWrite:
		return "RAW"
	case WriteAfterWrite:
		return "WAW"
	case WriteAfterRead:
		return "WAR"
	default:
		panic(h)
	}
}

type Hazard struct {
	Type     HazardType
	Register RegisterType
}

func (ctx *Context) IsDataHazard2(runner InstructionRunner) (bool, int, []RegisterType) {
	var hazards []RegisterType
	for _, register := range runner.ReadRegisters() {
		if register == Zero {
			continue
		}
		if v, exists := ctx.PendingWriteRegisters[register]; exists && v > 0 {
			// An instruction needs to read from a register that was updated
			hazards = append(hazards, register)
		}
	}
	if len(hazards) == 0 {
		return false, 0, nil
	}
	return true, len(hazards), hazards
}

type Execution struct {
	RegisterChange bool
	Register       RegisterType
	RegisterValue  int32
	MemoryChange   bool
	MemoryChanges  map[int32]int8
	NextPc         int32
	PcChange       bool
	Return         bool
}
