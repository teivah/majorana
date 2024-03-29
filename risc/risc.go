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

type InstructionType uint64

const (
	Add InstructionType = iota
	Addi
	And
	Andi
	Auipc
	Beq
	Bge
	Bgeu
	Blt
	Bltu
	Bne
	Div
	Jal
	Jalr
	Lui
	Lb
	Lh
	Lw
	Nop
	Mul
	Or
	Ori
	Rem
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

var CyclesPerInstruction = map[InstructionType]int{
	Add:   1,
	Addi:  1,
	And:   1,
	Andi:  1,
	Auipc: 1,
	Beq:   1,
	Bge:   1,
	Bgeu:  1,
	Blt:   1,
	Bltu:  1,
	Bne:   1,
	Div:   1,
	Jal:   1,
	Jalr:  1,
	Lui:   1,
	Lb:    50,
	Lh:    50,
	Lw:    50,
	Nop:   1,
	Mul:   1,
	Or:    1,
	Ori:   1,
	Rem:   1,
	Sb:    50,
	Sh:    50,
	Sll:   1,
	Slli:  1,
	Slt:   1,
	Sltu:  1,
	Slti:  1,
	Sra:   1,
	Srai:  1,
	Srl:   1,
	Srli:  1,
	Sub:   1,
	Sw:    50,
	Xor:   1,
	Xori:  1,
}

func IsWriteBack(ins InstructionType) bool {
	switch ins {
	case Sb, Sw, Sh:
		return false
	}
	return true
}

func IsJump(ins InstructionType) bool {
	switch ins {
	case Jal, Jalr:
		return true
	}
	return false
}

func IsConditionalBranching(ins InstructionType) bool {
	switch ins {
	case Beq, Bne, Blt, Bge, Bgeu:
		return true
	}
	return false
}

func IsRegisterChange(register RegisterType, value int32) (RegisterType, int32) {
	if register == Zero {
		return Zero, 0
	}
	return register, value
}
