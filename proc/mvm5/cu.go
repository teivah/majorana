package mvm5

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type controlUnit struct {
	inBus   *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus  *comp.BufferedBus[risc.InstructionRunnerPc]
	pending *risc.InstructionRunnerPc
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:  inBus,
		outBus: outBus,
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	if !u.outBus.CanAdd() {
		return
	}

	if u.pending != nil {
		u.outBus.Add(*u.pending, cycle)
		u.pending = nil
		return
	}

	remaining := u.outBus.RemainingToAdd()
	var runners []risc.InstructionRunnerPc
	for i := 0; i < remaining; i++ {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}
		if ctx.ContainWrittenRegisters(runner.Runner.ReadRegisters()) {
			u.inBus.Revert(runner, cycle)
			break
		}
		if ctx.Debug {
			fmt.Printf("\tCU: Adding runner %s at %d\n", runner.Runner.InstructionType(), runner.Pc/4)
		}
		runners = append(runners, runner)
	}

	if len(runners) == 1 {
		u.outBus.Add(runners[0], cycle)
		return
	}

	readRegs := make(map[risc.RegisterType]bool)
	writeRegs := make(map[risc.RegisterType]bool)
	for _, runner := range runners {
		for _, reg := range runner.Runner.ReadRegisters() {
			readRegs[reg] = true
		}
		for _, reg := range runner.Runner.WriteRegisters() {
			writeRegs[reg] = true
		}
	}

	if !isHazard(runners, readRegs, writeRegs) {
		for _, runner := range runners {
			u.outBus.Add(runner, cycle)
		}
	} else {
		u.outBus.Add(runners[0], cycle)
		for i := 1; i < len(runners); i++ {
			// TODO Logic only works if buffer == 2
			//u.outBus.Add(runners[i], cycle+1)
			u.pending = &runners[i]
			return
		}
	}
}

func isHazard(runners []risc.InstructionRunnerPc, readRegs, writeRegs map[risc.RegisterType]bool) bool {
	// Data hazard
	for reg := range writeRegs {
		if readRegs[reg] {
			return true
		}
	}
	for reg := range readRegs {
		if writeRegs[reg] {
			return true
		}
	}

	// Control hazard
	for _, runner := range runners {
		if risc.IsJump(runner.Runner.InstructionType()) ||
			risc.IsConditionalBranching(runner.Runner.InstructionType()) {
			return true
		}
	}
	return false
}

func (u *controlUnit) flush() {
}

func (u *controlUnit) isEmpty() bool {
	return true
}
