package risc

type RegisterType uint64

const (
	Zero RegisterType = iota
	Ra
	Sp
	Gp
	Tp
	T0
	T1
	T2
	S0
	S1
	A0
	A1
	A2
	A3
	A4
	A5
	A6
	A7
	S2
	S3
	S4
	S5
	S6
	S7
	S8
	S9
	S10
	S11
	T3
	T4
	T5
	T6
)

func (reg RegisterType) String() string {
	switch reg {
	case Zero:
		return "Zero"
	case Ra:
		return "Ra"
	case Sp:
		return "Sp"
	case Gp:
		return "Gp"
	case Tp:
		return "Tp"
	case T0:
		return "T0"
	case T1:
		return "T1"
	case T2:
		return "T2"
	case S0:
		return "S0"
	case S1:
		return "S1"
	case A0:
		return "A0"
	case A1:
		return "A1"
	case A2:
		return "A2"
	case A3:
		return "A3"
	case A4:
		return "A4"
	case A5:
		return "A5"
	case A6:
		return "A6"
	case A7:
		return "A7"
	case S2:
		return "S2"
	case S3:
		return "S3"
	case S4:
		return "S4"
	case S5:
		return "S5"
	case S6:
		return "S6"
	case S7:
		return "S7"
	case S8:
		return "S8"
	case S9:
		return "S9"
	case S10:
		return "S10"
	case S11:
		return "S11"
	case T3:
		return "T3"
	case T4:
		return "T4"
	case T5:
		return "T5"
	case T6:
		return "T6"
	default:
		panic(reg)
	}
}

type InstructionType uint64

const (
	Add InstructionType = iota
	Addi
	And
	Andi
	Auipc
	Beq
	Beqz
	Bge
	Bgeu
	Blt
	Bltu
	Bne
	Div
	J
	Jal
	Jalr
	Lui
	Lb
	Lh
	Li
	Lw
	Nop
	Mul
	Mv
	Or
	Ori
	Rem
	Ret
	Sb
	Sh
	Sll
	Slli
	Slt
	Sltu
	Slti
	Sra
	Srai
	Srl
	Srli
	Sub
	Sw
	Xor
	Xori
)

func (ins InstructionType) String() string {
	switch ins {
	case Add:
		return "Add"
	case Addi:
		return "Addi"
	case And:
		return "And"
	case Andi:
		return "Andi"
	case Auipc:
		return "Auipc"
	case Beq:
		return "Beq"
	case Beqz:
		return "Beqz"
	case Bge:
		return "Bge"
	case Bgeu:
		return "Bgeu"
	case Blt:
		return "Blt"
	case Bltu:
		return "Bltu"
	case Bne:
		return "Bne"
	case Div:
		return "Div"
	case J:
		return "J"
	case Jal:
		return "Jal"
	case Jalr:
		return "Jalr"
	case Lui:
		return "Lui"
	case Lb:
		return "Lb"
	case Lh:
		return "Lh"
	case Li:
		return "Li"
	case Lw:
		return "Lw"
	case Nop:
		return "Nop"
	case Mul:
		return "Mul"
	case Mv:
		return "Mv"
	case Or:
		return "Or"
	case Ori:
		return "Ori"
	case Rem:
		return "Rem"
	case Ret:
		return "Ret"
	case Sb:
		return "Sb"
	case Sh:
		return "Sh"
	case Sll:
		return "Sll"
	case Slli:
		return "Slli"
	case Slt:
		return "Slt"
	case Sltu:
		return "Sltu"
	case Slti:
		return "Slti"
	case Sra:
		return "Sra"
	case Srai:
		return "Srai"
	case Srl:
		return "Srl"
	case Srli:
		return "Srli"
	case Sub:
		return "Sub"
	case Sw:
		return "Sw"
	case Xor:
		return "Xor"
	case Xori:
		return "Xori"
	default:
		panic(ins)
	}
}

func (ins InstructionType) Cycles() int {
	switch ins {
	case Add:
		return 1
	case Addi:
		return 1
	case And:
		return 1
	case Andi:
		return 1
	case Auipc:
		return 1
	case Beq:
		return 1
	case Beqz:
		return 1
	case Bge:
		return 1
	case Bgeu:
		return 1
	case Blt:
		return 1
	case Bltu:
		return 1
	case Bne:
		return 1
	case Div:
		return 1
	case J:
		return 1
	case Jal:
		return 1
	case Jalr:
		return 1
	case Lui:
		return 1
	case Lb:
		return 50
	case Lh:
		return 50
	case Li:
		return 1
	case Lw:
		return 50
	case Nop:
		return 1
	case Mul:
		return 1
	case Mv:
		return 1
	case Or:
		return 1
	case Ori:
		return 1
	case Rem:
		return 1
	case Ret:
		return 1
	case Sb:
		// Write back
		return 1
	case Sh:
		// Write back
		return 1
	case Sll:
		return 1
	case Slli:
		return 1
	case Slt:
		return 1
	case Sltu:
		return 1
	case Slti:
		return 1
	case Sra:
		return 1
	case Srai:
		return 1
	case Srl:
		return 1
	case Srli:
		return 1
	case Sub:
		return 1
	case Sw:
		// Write back
		return 1
	case Xor:
		return 1
	case Xori:
		return 1
	default:
		panic(ins)
	}
}

func (ins InstructionType) IsWriteBack() bool {
	switch ins {
	case Sb, Sw, Sh:
		return false
	}
	return true
}

func (ins InstructionType) IsUnconditionalBranch() bool {
	switch ins {
	case J, Jal, Jalr:
		return true
	}
	return false
}

func (ins InstructionType) IsConditionalBranch() bool {
	switch ins {
	case Beq, Bne, Blt, Bge, Bgeu:
		return true
	}
	return false
}

func (ins InstructionType) IsBranch() bool {
	return ins.IsUnconditionalBranch() || ins.IsConditionalBranch()
}

func IsRegisterChange(register RegisterType, value int32) (RegisterType, int32) {
	if register == Zero {
		return Zero, 0
	}
	return register, value
}
