package mvm5

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type controlUnit struct {
	inBus  *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus *comp.BufferedBus[risc.InstructionRunnerPc]
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

	remaining := u.outBus.RemainingToAdd()
	var runners []risc.InstructionRunnerPc
	for i := 0; i < remaining; i++ {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}
		runners = append(runners, runner)
	}

	if len(runners) == 1 {
		u.outBus.Add(runners[0], cycle)
		return
	}

	hazard := false
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

	// Data hazard
	for reg := range writeRegs {
		if readRegs[reg] {
			hazard = true
		}
	}
	for reg := range readRegs {
		if writeRegs[reg] {
			hazard = true
		}
	}

	if !hazard {
		for _, runner := range runners {
			u.outBus.Add(runner, cycle)
		}
	} else {
		u.outBus.Add(runners[0], cycle)
		for i := 1; i < len(runners); i++ {
			// TODO Logic only works if buffer == 2
			u.outBus.Add(runners[i], cycle)
		}
	}
}

func (u *controlUnit) flush() {
}

func (u *controlUnit) isEmpty() bool {
	return true
}
