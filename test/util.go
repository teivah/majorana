package test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teivah/majorana/risc"
)

func RunAssert(t *testing.T, initRegisters map[risc.RegisterType]int32, memoryBytes int, initMemory map[int]int8, instructions string, assertionsRegisters map[risc.RegisterType]int32, assertionsMemory map[int]int8) {
	app, err := risc.Parse(instructions, false)
	require.NoError(t, err)
	r := risc.NewRunner(app, memoryBytes)
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

func ReadFile(t *testing.T, filename string) string {
	content, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	return string(content)
}
