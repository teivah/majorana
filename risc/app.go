package risc

type Application struct {
	Instructions []InstructionRunner
	Labels       map[string]int32
}

type Context struct {
	Registers     map[RegisterType]int32
	ReadRegisters map[RegisterType]int
	Memory        []int8
	Debug         bool
}

func NewContext(debug bool, memoryBytes int) *Context {
	return &Context{
		Registers:     make(map[RegisterType]int32),
		ReadRegisters: make(map[RegisterType]int),
		Memory:        make([]int8, memoryBytes),
		Debug:         debug,
	}
}

func (ctx *Context) WriteRegister(exe Execution) {
	ctx.Registers[exe.Register] = exe.RegisterValue
}

func (ctx *Context) WriteMemory(exe Execution) {
	for k, v := range exe.MemoryChanges {
		ctx.Memory[k] = v
	}
}

func (ctx *Context) AddWriteRegisters(registers []RegisterType) {
	for _, register := range registers {
		ctx.ReadRegisters[register]++
	}
}

func (ctx *Context) DeleteWriteRegisters(registers []RegisterType) {
	for _, register := range registers {
		ctx.ReadRegisters[register]--
		if ctx.Registers[register] <= 0 {
			delete(ctx.Registers, register)
		}
	}
}

func (ctx *Context) ContainWrittenRegisters(registers []RegisterType) bool {
	for _, register := range registers {
		if register == Zero {
			continue
		}
		if v, exists := ctx.ReadRegisters[register]; exists && v > 0 {
			return true
		}
	}
	return false
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
