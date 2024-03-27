package risc

type Application struct {
	Instructions []InstructionRunner
	Labels       map[string]int32
}

type Context struct {
	Registers     map[RegisterType]int32
	ReadRegisters map[RegisterType]struct{}
	Memory        []int8
	Pc            int32
}

func NewContext(memoryBytes int) *Context {
	return &Context{
		Registers:     make(map[RegisterType]int32),
		ReadRegisters: make(map[RegisterType]struct{}),
		Memory:        make([]int8, memoryBytes),
	}
}

func (ctx *Context) Write(exe Execution) {
	ctx.Registers[exe.Register] = exe.Value
}

func (ctx *Context) AddWriteRegisters(registers []RegisterType) {
	for _, register := range registers {
		ctx.ReadRegisters[register] = struct{}{}
	}
}

func (ctx *Context) DeleteWriteRegisters(registers []RegisterType) {
	for _, register := range registers {
		delete(ctx.ReadRegisters, register)
	}
}

func (ctx *Context) ContainWrittenRegisters(registers []RegisterType) bool {
	for _, register := range registers {
		if _, exists := ctx.ReadRegisters[register]; exists {
			return true
		}
	}
	return false
}

type Execution struct {
	Register RegisterType
	Value    int32
	Pc       int32
}

func newExecution(register RegisterType, value, pc int32) Execution {
	return Execution{
		Register: register,
		Value:    value,
		Pc:       pc,
	}
}

func pc(pc int32) Execution {
	return Execution{
		Register: Zero,
		Value:    0,
		Pc:       pc,
	}
}
