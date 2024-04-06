package risc

import "fmt"

type InstructionRunnerPc struct {
	Runner InstructionRunner
	Pc     int32

	Forwarder       chan<- int32
	Receiver        <-chan int32
	ForwardRegister RegisterType
}

type Forward struct {
	Register RegisterType
	Value    int32
}

func registerRead(ctx *Context, forward Forward, reg RegisterType) int32 {
	if reg == forward.Register {
		return forward.Value
	}
	return ctx.Registers[reg]
}

type InstructionRunner interface {
	Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error)
	InstructionType() InstructionType
	ReadRegisters() []RegisterType
	WriteRegisters() []RegisterType
	Forward(forward Forward)
	MemoryRead(ctx *Context) []int32
}

type add struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *add) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1+rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *add) InstructionType() InstructionType {
	return Add
}

func (op *add) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *add) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *add) Forward(forward Forward) {
	op.forward = forward
}

func (op *add) MemoryRead(ctx *Context) []int32 {
	return nil
}

type addi struct {
	imm     int32
	rd      RegisterType
	rs      RegisterType
	forward Forward
}

func (op *addi) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, rs+op.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *addi) InstructionType() InstructionType {
	return Addi
}

func (op *addi) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *addi) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *addi) Forward(forward Forward) {
	op.forward = forward
}

func (op *addi) MemoryRead(ctx *Context) []int32 {
	return nil
}

type and struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *and) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1&rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *and) InstructionType() InstructionType {
	return And
}

func (op *and) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *and) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *and) Forward(forward Forward) {
	op.forward = forward
}

func (op *and) MemoryRead(ctx *Context) []int32 {
	return nil
}

type andi struct {
	imm     int32
	rd      RegisterType
	rs      RegisterType
	forward Forward
}

func (op *andi) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, rs&op.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *andi) InstructionType() InstructionType {
	return Andi
}

func (op *andi) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *andi) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *andi) Forward(forward Forward) {
	op.forward = forward
}

func (op *andi) MemoryRead(ctx *Context) []int32 {
	return nil
}

type auipc struct {
	rd  RegisterType
	imm int32
}

func (op *auipc) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	register, value := IsRegisterChange(op.rd, pc+(op.imm<<12))
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *auipc) InstructionType() InstructionType {
	return Auipc
}

func (op *auipc) ReadRegisters() []RegisterType {
	return nil
}

func (op *auipc) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *auipc) Forward(forward Forward) {
}

func (op *auipc) MemoryRead(ctx *Context) []int32 {
	return nil
}

type beq struct {
	rs1     RegisterType
	rs2     RegisterType
	label   string
	forward Forward
}

func (op *beq) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs1 == rs2 {
		addr, ok := labels[op.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", op.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (op *beq) InstructionType() InstructionType {
	return Beq
}

func (op *beq) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *beq) WriteRegisters() []RegisterType {
	return nil
}

func (op *beq) Forward(forward Forward) {
	op.forward = forward
}

func (op *beq) MemoryRead(ctx *Context) []int32 {
	return nil
}

type beqz struct {
	rs      RegisterType
	label   string
	forward Forward
}

func (op *beqz) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	if rs == 0 {
		addr, ok := labels[op.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", op.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (op *beqz) InstructionType() InstructionType {
	return Beqz
}

func (op *beqz) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs, op.rs}
}

func (op *beqz) WriteRegisters() []RegisterType {
	return nil
}

func (op *beqz) Forward(forward Forward) {
	op.forward = forward
}

func (op *beqz) MemoryRead(ctx *Context) []int32 {
	return nil
}

type bge struct {
	rs1     RegisterType
	rs2     RegisterType
	label   string
	forward Forward
}

func (op *bge) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs1 >= rs2 {
		addr, ok := labels[op.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", op.label)
		}
		if ctx.Debug {
			fmt.Printf("\t\tRun: bge true %d\n", addr/4)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	if ctx.Debug {
		fmt.Printf("\t\tRun: bge false\n")
	}
	return Execution{}, nil
}

func (op *bge) InstructionType() InstructionType {
	return Bge
}

func (op *bge) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *bge) WriteRegisters() []RegisterType {
	return nil
}

func (op *bge) Forward(forward Forward) {
	op.forward = forward
}

func (op *bge) MemoryRead(ctx *Context) []int32 {
	return nil
}

type bgeu struct {
	rs1     RegisterType
	rs2     RegisterType
	label   string
	forward Forward
}

func (op *bgeu) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs1 >= rs2 {
		addr, ok := labels[op.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", op.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (op *bgeu) InstructionType() InstructionType {
	return Bgeu
}

func (op *bgeu) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *bgeu) WriteRegisters() []RegisterType {
	return nil
}

func (op *bgeu) Forward(forward Forward) {
	op.forward = forward
}

func (op bgeu) MemoryRead(ctx *Context) []int32 {
	return nil
}

type blt struct {
	rs1     RegisterType
	rs2     RegisterType
	label   string
	forward Forward
}

func (op *blt) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs1 < rs2 {
		addr, ok := labels[op.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", op.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (op *blt) InstructionType() InstructionType {
	return Blt
}

func (op *blt) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *blt) WriteRegisters() []RegisterType {
	return nil
}

func (op *blt) Forward(forward Forward) {
	op.forward = forward
}

func (op *blt) MemoryRead(ctx *Context) []int32 {
	return nil
}

type bltu struct {
	rs1     RegisterType
	rs2     RegisterType
	label   string
	forward Forward
}

func (op *bltu) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs1 < rs2 {
		addr, ok := labels[op.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", op.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (op *bltu) InstructionType() InstructionType {
	return Bltu
}

func (op *bltu) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *bltu) WriteRegisters() []RegisterType {
	return nil
}

func (op *bltu) Forward(forward Forward) {
	op.forward = forward
}

func (op *bltu) MemoryRead(ctx *Context) []int32 {
	return nil
}

type bne struct {
	rs1     RegisterType
	rs2     RegisterType
	label   string
	forward Forward
}

func (op *bne) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs1 != rs2 {
		addr, ok := labels[op.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", op.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (op *bne) InstructionType() InstructionType {
	return Bne
}

func (op *bne) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *bne) WriteRegisters() []RegisterType {
	return nil
}

func (op *bne) Forward(forward Forward) {
	op.forward = forward
}

func (op *bne) MemoryRead(ctx *Context) []int32 {
	return nil
}

type div struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *div) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs2 == 0 {
		return Execution{}, fmt.Errorf("division by zero")
	}
	register, value := IsRegisterChange(op.rd, rs1/rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *div) InstructionType() InstructionType {
	return Div
}

func (op *div) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *div) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *div) Forward(forward Forward) {
	op.forward = forward
}

func (op *div) MemoryRead(ctx *Context) []int32 {
	return nil
}

type j struct {
	label string
}

func (op *j) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	addr, ok := labels[op.label]
	if !ok {
		return Execution{}, fmt.Errorf("label %s does not exist", op.label)
	}
	return Execution{
		NextPc:   addr,
		PcChange: true,
	}, nil
}

func (op *j) InstructionType() InstructionType {
	return J
}

func (op *j) ReadRegisters() []RegisterType {
	return nil
}

func (op *j) WriteRegisters() []RegisterType {
	return nil
}

func (op *j) Forward(forward Forward) {
}

func (op *j) MemoryRead(ctx *Context) []int32 {
	return nil
}

type jal struct {
	label   string
	rd      RegisterType
	forward Forward
}

func (op *jal) Run(ctx *Context, labels map[string]int32, pc int32, memory []int8) (Execution, error) {
	addr, ok := labels[op.label]
	if !ok {
		return Execution{}, fmt.Errorf("label %s does not exist", op.label)
	}
	// TODO Shouldn't be a direct write
	ctx.Registers[Ra] = pc
	register, value := IsRegisterChange(op.rd, pc+4)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
		NextPc:         addr,
		PcChange:       true,
	}, nil
}

func (op *jal) InstructionType() InstructionType {
	return Jal
}

func (op *jal) ReadRegisters() []RegisterType {
	return nil
}

func (op *jal) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *jal) Forward(forward Forward) {
	op.forward = forward
}

func (op *jal) MemoryRead(ctx *Context) []int32 {
	return nil
}

type jalr struct {
	rd      RegisterType
	rs      RegisterType
	imm     int32
	forward Forward
}

func (op *jalr) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, pc+4)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
		NextPc:         rs + op.imm,
		PcChange:       true,
	}, nil
}

func (op *jalr) InstructionType() InstructionType {
	return Jalr
}

func (op *jalr) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *jalr) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *jalr) Forward(forward Forward) {
	op.forward = forward
}

func (op *jalr) MemoryRead(ctx *Context) []int32 {
	return nil
}

type lui struct {
	rd  RegisterType
	imm int32
}

func (op *lui) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	register, value := IsRegisterChange(op.rd, op.imm<<12)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *lui) InstructionType() InstructionType {
	return Lui
}

func (op *lui) ReadRegisters() []RegisterType {
	return nil
}

func (op *lui) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *lui) Forward(forward Forward) {
}

func (op *lui) MemoryRead(ctx *Context) []int32 {
	return nil
}

type lb struct {
	rd      RegisterType
	offset  int32
	rs      RegisterType
	forward Forward
}

func (op *lb) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	n := memory[0]
	register, value := IsRegisterChange(op.rd, int32(n))
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *lb) InstructionType() InstructionType {
	return Lb
}

func (op *lb) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *lb) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *lb) Forward(forward Forward) {
	op.forward = forward
}

func (op *lb) MemoryRead(ctx *Context) []int32 {
	rs := registerRead(ctx, op.forward, op.rs)
	return []int32{rs + op.offset}
}

type lh struct {
	rd      RegisterType
	offset  int32
	rs      RegisterType
	forward Forward
}

func (op *lh) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	n := I32FromBytes(memory[0], memory[1], 0, 0)
	register, value := IsRegisterChange(op.rd, n)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *lh) InstructionType() InstructionType {
	return Lh
}

func (op *lh) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs, op.rd}
}

func (op *lh) WriteRegisters() []RegisterType {
	return nil
}

func (op *lh) Forward(forward Forward) {
	op.forward = forward
}

func (op *lh) MemoryRead(ctx *Context) []int32 {
	rs := registerRead(ctx, op.forward, op.rs)
	idx := rs + op.offset
	return []int32{idx, idx + 1}
}

type li struct {
	rd  RegisterType
	imm int32
}

func (op *li) Run(ctx *Context, _ map[string]int32, _ int32, memory []int8) (Execution, error) {
	register, value := IsRegisterChange(op.rd, op.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *li) InstructionType() InstructionType {
	return Li
}

func (op *li) ReadRegisters() []RegisterType {
	return nil
}

func (op *li) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *li) Forward(forward Forward) {
}

func (op *li) MemoryRead(ctx *Context) []int32 {
	return nil
}

type lw struct {
	rd      RegisterType
	offset  int32
	rs      RegisterType
	forward Forward
}

func (op *lw) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	n := I32FromBytes(memory[0], memory[1], memory[2], memory[3])
	register, value := IsRegisterChange(op.rd, n)
	if ctx.Debug {
		fmt.Printf("\t\tRun: Lw %s %d\n", register, value)
	}
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *lw) InstructionType() InstructionType {
	return Lw
}

func (op *lw) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *lw) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *lw) Forward(forward Forward) {
	op.forward = forward
}

func (op *lw) MemoryRead(ctx *Context) []int32 {
	rs := registerRead(ctx, op.forward, op.rs)
	idx := rs + op.offset
	return []int32{idx, idx + 1, idx + 2, idx + 3}
}

type nop struct{}

func (op *nop) Run(_ *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	return Execution{}, nil
}

func (op *nop) InstructionType() InstructionType {
	return Nop
}

func (op *nop) ReadRegisters() []RegisterType {
	return nil
}

func (op *nop) WriteRegisters() []RegisterType {
	return nil
}

func (op *nop) Forward(forward Forward) {
}

func (op *nop) MemoryRead(ctx *Context) []int32 {
	return nil
}

type mul struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *mul) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1*rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *mul) InstructionType() InstructionType {
	return Mul
}

func (op *mul) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *mul) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *mul) Forward(forward Forward) {
	op.forward = forward
}

func (op *mul) MemoryRead(ctx *Context) []int32 {
	return nil
}

type mv struct {
	rd      RegisterType
	rs      RegisterType
	forward Forward
}

func (op *mv) Run(ctx *Context, _ map[string]int32, _ int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, rs)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *mv) InstructionType() InstructionType {
	return Mv
}

func (op *mv) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *mv) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *mv) Forward(forward Forward) {
	op.forward = forward
}

func (op *mv) MemoryRead(ctx *Context) []int32 {
	return nil
}

type or struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *or) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1|rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *or) InstructionType() InstructionType {
	return Or
}

func (op *or) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *or) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *or) Forward(forward Forward) {
	op.forward = forward
}

func (op *or) MemoryRead(ctx *Context) []int32 {
	return nil
}

type ori struct {
	imm     int32
	rd      RegisterType
	rs      RegisterType
	forward Forward
}

func (op *ori) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, rs|op.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *ori) InstructionType() InstructionType {
	return Ori
}

func (op *ori) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *ori) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *ori) Forward(forward Forward) {
	op.forward = forward
}

func (op *ori) MemoryRead(ctx *Context) []int32 {
	return nil
}

type rem struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *rem) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if ctx.Debug {
		fmt.Printf("\t\tRun: Rem %d %d\n", rs1, rs2)
	}
	register, value := IsRegisterChange(op.rd, rs1%rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *rem) InstructionType() InstructionType {
	return Rem
}

func (op *rem) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *rem) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *rem) Forward(forward Forward) {
	op.forward = forward
}

func (op *rem) MemoryRead(ctx *Context) []int32 {
	return nil
}

type ret struct{}

func (op *ret) Run(_ *Context, _ map[string]int32, _ int32, memory []int8) (Execution, error) {
	return Execution{Return: true}, nil
}

func (op *ret) InstructionType() InstructionType {
	return Ret
}

func (op *ret) ReadRegisters() []RegisterType {
	return nil
}

func (op *ret) WriteRegisters() []RegisterType {
	return nil
}

func (op *ret) Forward(forward Forward) {
}

func (op *ret) MemoryRead(ctx *Context) []int32 {
	return nil
}

type sb struct {
	rs2     RegisterType
	offset  int32
	rs1     RegisterType
	forward Forward
}

func (op *sb) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	idx := rs1 + op.offset
	n := rs2
	return Execution{
		MemoryChange:  true,
		MemoryChanges: map[int32]int8{idx: int8(n)},
	}, nil
}

func (op *sb) InstructionType() InstructionType {
	return Sb
}

func (op *sb) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *sb) WriteRegisters() []RegisterType {
	return nil
}

func (op *sb) Forward(forward Forward) {
	op.forward = forward
}

func (op *sb) MemoryRead(ctx *Context) []int32 {
	return nil
}

type sh struct {
	rs2     RegisterType
	offset  int32
	rs1     RegisterType
	forward Forward
}

func (op *sh) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	idx := rs1 + op.offset
	n := rs2
	bytes := BytesFromLowBits(n)
	return Execution{
		MemoryChange: true,
		MemoryChanges: map[int32]int8{
			idx:     bytes[0],
			idx + 1: bytes[1],
		},
	}, nil
}

func (op *sh) InstructionType() InstructionType {
	return Sh
}

func (op *sh) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *sh) WriteRegisters() []RegisterType {
	return nil
}

func (op *sh) Forward(forward Forward) {
	op.forward = forward
}

func (op *sh) MemoryRead(ctx *Context) []int32 {
	return nil
}

type sll struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *sll) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1<<uint(rs2))
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *sll) InstructionType() InstructionType {
	return Sll
}

func (op *sll) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *sll) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *sll) Forward(forward Forward) {
	op.forward = forward
}

func (op *sll) MemoryRead(ctx *Context) []int32 {
	return nil
}

type slli struct {
	rd      RegisterType
	rs      RegisterType
	imm     int32
	forward Forward
}

func (op *slli) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, rs<<uint(op.imm))
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *slli) InstructionType() InstructionType {
	return Slli
}

func (op *slli) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *slli) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *slli) Forward(forward Forward) {
	op.forward = forward
}

func (op *slli) MemoryRead(ctx *Context) []int32 {
	return nil
}

type slt struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *slt) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	var register RegisterType
	var value int32
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs1 < rs2 {
		register, value = IsRegisterChange(op.rd, 1)
	} else {
		register, value = IsRegisterChange(op.rd, 0)
	}
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *slt) InstructionType() InstructionType {
	return Slt
}

func (op *slt) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *slt) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *slt) Forward(forward Forward) {
	op.forward = forward
}

func (op *slt) MemoryRead(ctx *Context) []int32 {
	return nil
}

type sltu struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *sltu) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	var register RegisterType
	var value int32
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	if rs1 < rs2 {
		register, value = IsRegisterChange(op.rd, 1)
	} else {
		register, value = IsRegisterChange(op.rd, 0)
	}
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *sltu) InstructionType() InstructionType {
	return Sltu
}

func (op *sltu) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *sltu) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *sltu) Forward(forward Forward) {
	op.forward = forward
}

func (op *sltu) MemoryRead(ctx *Context) []int32 {
	return nil
}

type slti struct {
	rd      RegisterType
	rs      RegisterType
	imm     int32
	forward Forward
}

func (op *slti) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	var register RegisterType
	var value int32
	rs := registerRead(ctx, op.forward, op.rs)
	if rs < op.imm {
		register, value = IsRegisterChange(op.rd, 1)
	} else {
		register, value = IsRegisterChange(op.rd, 0)
	}
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *slti) InstructionType() InstructionType {
	return Slti
}

func (op *slti) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *slti) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *slti) Forward(forward Forward) {
	op.forward = forward
}

func (op *slti) MemoryRead(ctx *Context) []int32 {
	return nil
}

type sra struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *sra) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1>>rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *sra) InstructionType() InstructionType {
	return Sra
}

func (op *sra) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *sra) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *sra) Forward(forward Forward) {
	op.forward = forward
}

func (op *sra) MemoryRead(ctx *Context) []int32 {
	return nil
}

type srai struct {
	rd      RegisterType
	rs      RegisterType
	imm     int32
	forward Forward
}

func (op *srai) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, rs>>op.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *srai) InstructionType() InstructionType {
	return Srai
}

func (op *srai) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *srai) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *srai) Forward(forward Forward) {
	op.forward = forward
}

func (op *srai) MemoryRead(ctx *Context) []int32 {
	return nil
}

type srl struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *srl) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1>>rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *srl) InstructionType() InstructionType {
	return Srl
}

func (op *srl) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *srl) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *srl) Forward(forward Forward) {
	op.forward = forward
}

func (op *srl) MemoryRead(ctx *Context) []int32 {
	return nil
}

type srli struct {
	rd      RegisterType
	rs      RegisterType
	imm     int32
	forward Forward
}

func (op *srli) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, rs>>op.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *srli) InstructionType() InstructionType {
	return Srli
}

func (op *srli) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *srli) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *srli) Forward(forward Forward) {
	op.forward = forward
}

func (op *srli) MemoryRead(ctx *Context) []int32 {
	return nil
}

type sub struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *sub) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1-rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *sub) InstructionType() InstructionType {
	return Sub
}

func (op *sub) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *sub) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *sub) Forward(forward Forward) {
	op.forward = forward
}

func (op *sub) MemoryRead(ctx *Context) []int32 {
	return nil
}

type sw struct {
	rs2     RegisterType
	offset  int32
	rs1     RegisterType
	forward Forward
}

func (op *sw) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	idx := rs1 + op.offset
	n := rs2
	bytes := BytesFromLowBits(n)
	if ctx.Debug {
		fmt.Printf("\t\tRun: Sw %d to %d\n", idx, n)
	}
	return Execution{
		MemoryChange: true,
		MemoryChanges: map[int32]int8{
			idx:     bytes[0],
			idx + 1: bytes[1],
			idx + 2: bytes[2],
			idx + 3: bytes[3],
		},
	}, nil
}

func (op *sw) InstructionType() InstructionType {
	return Sw
}

func (op *sw) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *sw) WriteRegisters() []RegisterType {
	return nil
}

func (op *sw) Forward(forward Forward) {
	op.forward = forward
}

func (op *sw) MemoryRead(ctx *Context) []int32 {
	return nil
}

type xor struct {
	rd      RegisterType
	rs1     RegisterType
	rs2     RegisterType
	forward Forward
}

func (op *xor) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs1 := registerRead(ctx, op.forward, op.rs1)
	rs2 := registerRead(ctx, op.forward, op.rs2)
	register, value := IsRegisterChange(op.rd, rs1^rs2)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *xor) InstructionType() InstructionType {
	return Xor
}

func (op *xor) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs1, op.rs2}
}

func (op *xor) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *xor) Forward(forward Forward) {
	op.forward = forward
}

func (op *xor) MemoryRead(ctx *Context) []int32 {
	return nil
}

type xori struct {
	imm     int32
	rd      RegisterType
	rs      RegisterType
	forward Forward
}

func (op *xori) Run(ctx *Context, _ map[string]int32, pc int32, memory []int8) (Execution, error) {
	rs := registerRead(ctx, op.forward, op.rs)
	register, value := IsRegisterChange(op.rd, rs^op.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (op *xori) InstructionType() InstructionType {
	return Xori
}

func (op *xori) ReadRegisters() []RegisterType {
	return []RegisterType{op.rs}
}

func (op *xori) WriteRegisters() []RegisterType {
	return []RegisterType{op.rd}
}

func (op *xori) Forward(forward Forward) {
	op.forward = forward
}

func (op *xori) MemoryRead(ctx *Context) []int32 {
	return nil
}
