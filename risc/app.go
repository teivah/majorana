package risc

import "fmt"

type ExecutionContext struct {
	Pc              int32
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
	Registers             map[RegisterType]int32
	PendingWriteRegisters map[RegisterType]int
	PendingReadRegisters  map[RegisterType]int
	Memory                []int8
	Debug                 bool
}

func NewContext(debug bool, memoryBytes int) *Context {
	return &Context{
		Registers:             make(map[RegisterType]int32),
		PendingWriteRegisters: make(map[RegisterType]int),
		PendingReadRegisters:  make(map[RegisterType]int),
		Memory:                make([]int8, memoryBytes),
		Debug:                 debug,
	}
}

func (ctx *Context) Flush() {
	ctx.PendingWriteRegisters = make(map[RegisterType]int)
	ctx.PendingReadRegisters = make(map[RegisterType]int)
}

func (ctx *Context) WriteRegister(exe Execution) {
	ctx.Registers[exe.Register] = exe.RegisterValue
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
