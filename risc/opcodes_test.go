package risc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runAssert(t *testing.T, initRegisters map[RegisterType]int32, memoryBytes int, initMemory map[int]int8, instructions string, assertionsRegisters map[RegisterType]int32, assertionsMemory map[int]int8) {
	app, err := Parse(instructions)
	require.NoError(t, err)
	r := NewRunner(app, memoryBytes)
	for k, v := range initRegisters {
		r.Ctx.Registers[k] = v
	}
	for k, v := range initMemory {
		r.Ctx.Memory[k] = v
	}

	err = r.Run()
	require.NoError(t, err)

	for k, v := range assertionsRegisters {
		assert.Equal(t, v, r.Ctx.Registers[k], "register")
	}
	for k, v := range assertionsMemory {
		assert.Equal(t, v, r.Ctx.Memory[k], "memory")
	}
}

func TestAdd(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 1, T2: 2}, 0, map[int]int8{},
		"add t0, t1, t2", map[RegisterType]int32{T0: 3}, map[int]int8{})
}

func TestAddi(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 1}, 0, map[int]int8{},
		"addi t0, t1, 1", map[RegisterType]int32{T0: 2}, map[int]int8{})
}

func TestAnd(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 1, T2: 3}, 0, map[int]int8{},
		"and t0, t1, t2", map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestAndi(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 1}, 0, map[int]int8{},
		"andi t0, t1, 3", map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestAuipc(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`auipc t0, 0
auipc t0, 0
auipc t0, 0`, map[RegisterType]int32{T0: 8}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`auipc t0, 1
auipc t0, 1
auipc t0, 1`, map[RegisterType]int32{T0: 4104}, map[int]int8{})
}

func TestBeq(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`beq t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 0, T1: 1}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{T0: 1}, 0, map[int]int8{},
		`beq t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 2, T1: 1}, map[int]int8{})
}

func TestBge(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`bge t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 0, T1: 1}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{T1: 10}, 0, map[int]int8{},
		`bge t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 2, T1: 1}, map[int]int8{})
}

func TestBgeu(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`bgeu t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 0, T1: 1}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{T1: 10}, 0, map[int]int8{},
		`bgeu t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 2, T1: 1}, map[int]int8{})
}

func TestBlt(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`blt t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 2, T1: 1}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{T1: 10}, 0, map[int]int8{},
		`blt t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 0, T1: 1}, map[int]int8{})
}

func TestBltu(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`blt t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 2, T1: 1}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{T1: 10}, 0, map[int]int8{},
		`blt t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 0, T1: 1}, map[int]int8{})
}

func TestBne(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`bne t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 2, T1: 1}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{T0: 1}, 0, map[int]int8{},
		`bne t0, t1, foo
addi t0, zero, 2
foo:
addi t1, zero, 1`, map[RegisterType]int32{T0: 1, T1: 1}, map[int]int8{})
}

func TestDiv(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 4, T2: 2}, 0, map[int]int8{},
		`div t0, t1, t2`, map[RegisterType]int32{T0: 2}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{T1: 4, T2: 3}, 0, map[int]int8{},
		`div t0, t1, t2`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestJal(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`jal t0, foo
addi t1, zero, 1
foo:
addi t2, zero, 2`, map[RegisterType]int32{T0: 4, T1: 0, T2: 2}, map[int]int8{})
}

func TestJalr(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`addi t1, zero, 4
jalr t0, t1, 8
foo:
addi t2, zero, 2
addi t1, zero, 2`, map[RegisterType]int32{T0: 8, T1: 2, T2: 0}, map[int]int8{})
}

func TestLui(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`lui t0, 0`, map[RegisterType]int32{T0: 0}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`lui t0, 1`, map[RegisterType]int32{T0: 4096}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`lui t0, 3`, map[RegisterType]int32{T0: 12288}, map[int]int8{})
}

func TestLi(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		`li t0, 42`, map[RegisterType]int32{T0: 42}, map[int]int8{})
}

func TestMul(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 4, T2: 2}, 0, map[int]int8{},
		`mul t0, t1, t2`, map[RegisterType]int32{T0: 8}, map[int]int8{})
}

func TestOr(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 1, T2: 2}, 0, map[int]int8{},
		`or t0, t1, t2`, map[RegisterType]int32{T0: 3}, map[int]int8{})
}

func TestOri(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 1}, 0, map[int]int8{},
		`ori t0, t1, 2`, map[RegisterType]int32{T0: 3}, map[int]int8{})
}

func TestRem(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 4, T2: 2}, 0, map[int]int8{},
		`rem t0, t1, t2`, map[RegisterType]int32{T0: 0}, map[int]int8{})

	runAssert(t, map[RegisterType]int32{T1: 4, T2: 3}, 0, map[int]int8{},
		`rem t0, t1, t2`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestSll(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 1, T2: 2}, 0, map[int]int8{},
		`sll t0, t1, t2`, map[RegisterType]int32{T0: 4}, map[int]int8{})
}

func TestSlli(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 1}, 0, map[int]int8{},
		`slli t0, t1, 2`, map[RegisterType]int32{T0: 4}, map[int]int8{})
}

func TestSlt(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 2, T2: 3}, 0, map[int]int8{},
		`slt t0, t1, t2`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestSlti(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 2}, 0, map[int]int8{},
		`slti t0, t1, 5`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestSltu(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 2, T2: 3}, 0, map[int]int8{},
		`sltu t0, t1, t2`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestSra(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 2, T2: 1}, 0, map[int]int8{},
		`sra t0, t1, t2`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestSrai(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 2}, 0, map[int]int8{},
		`srai t0, t1, 1`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestSrl(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 4, T2: 2}, 0, map[int]int8{},
		`srl t0, t1, t2`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestSrli(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 4}, 0, map[int]int8{},
		`srli t0, t1, 2`, map[RegisterType]int32{T0: 1}, map[int]int8{})
}

func TestSub(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 10, T2: 6}, 0, map[int]int8{},
		`sub t0, t1, t2`, map[RegisterType]int32{T0: 4}, map[int]int8{})
}

func TestSbLb(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T0: 16, T1: 2}, 8, map[int]int8{},
		`sb t0, 2, t1
	lb t2, 2, t1`, map[RegisterType]int32{T2: 16}, map[int]int8{4: 16})

	runAssert(t, map[RegisterType]int32{T0: 2047, T1: 2}, 8, map[int]int8{},
		`sb t0, 2, t1
lb t2, 2, t1`, map[RegisterType]int32{T2: -1}, map[int]int8{4: -1})
}

func TestShLh(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T0: 64, T1: 2}, 8, map[int]int8{4: 1, 5: 1},
		`sh t0, 2, t1
lh t2, 2, t1`, map[RegisterType]int32{T2: 64}, map[int]int8{4: 64, 5: 0})

	runAssert(t, map[RegisterType]int32{T0: 2047, T1: 2}, 8, map[int]int8{4: 1, 5: 1},
		`sh t0, 2, t1
lh t2, 2, t1`, map[RegisterType]int32{T2: 2047}, map[int]int8{4: -1, 5: 7})
}

func TestSwLw(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T0: 258, T1: 2}, 8, map[int]int8{4: 1, 5: 1, 6: 1, 7: 1},
		`sw t0, 2, t1
lw t2, 2(t1)`, map[RegisterType]int32{T2: 258}, map[int]int8{4: 2, 5: 1, 6: 0, 7: 0})

	runAssert(t, map[RegisterType]int32{T0: 2047, T1: 2}, 8, map[int]int8{4: 1, 5: 1, 6: 1, 7: 1},
		`sw t0, 2, t1
lw t2, 2(t1)`, map[RegisterType]int32{T2: 2047}, map[int]int8{4: -1, 5: 7, 6: 0, 7: 0})
}

func TestXor(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 3, T2: 4}, 0, map[int]int8{},
		"xor t0, t1, t2", map[RegisterType]int32{T0: 7}, map[int]int8{})
}

func TestXori(t *testing.T) {
	runAssert(t, map[RegisterType]int32{T1: 3}, 0, map[int]int8{},
		"xori t0, t1, 4", map[RegisterType]int32{T0: 7}, map[int]int8{})
}

func TestZero(t *testing.T) {
	runAssert(t, map[RegisterType]int32{}, 0, map[int]int8{},
		"addi zero, zero, 1", map[RegisterType]int32{Zero: 0}, map[int]int8{})
}
