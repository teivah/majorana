package mvp6

import (
	"fmt"

	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/common/obs"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	pendingLength = 2
)

type controlUnit struct {
	inBus    *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus   *comp.BufferedBus[risc.InstructionRunnerPc]
	pendings *comp.Queue[risc.InstructionRunnerPc]

	pushed            *obs.Gauge
	blocked           *obs.Gauge
	total             int
	cantAdd           int
	blockedBranch     int
	blockedDataHazard int
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:    inBus,
		outBus:   outBus,
		pendings: comp.NewQueue[risc.InstructionRunnerPc](pendingLength),
		pushed:   &obs.Gauge{},
		blocked:  &obs.Gauge{},
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	pushed := 0
	defer func() {
		u.pushed.Push(pushed)
	}()
	if u.inBus.CanGet() {
		u.blocked.Push(1)
	} else {
		u.blocked.Push(0)
	}
	u.total++

	if !u.outBus.CanAdd() {
		u.cantAdd++
		log.Infou(ctx, "CU", "can't add")
		return
	}

	remaining := u.outBus.RemainingToAdd()
	var pushedRunners []risc.InstructionRunner

	defer func() {
		for _, runner := range pushedRunners {
			ctx.AddPendingRegisters(runner)
		}
	}()

	for elem := range u.pendings.Iterator() {
		pending := u.pendings.Value(elem)
		if pushed > 0 && pending.Runner.InstructionType().IsBranch() {
			u.blockedBranch++
			return
		}

		hazard, reason := ctx.IsDataHazard(pending.Runner)
		if !hazard {
			hazard, reason := isHazardWithPushedRunners(pushedRunners, pending.Runner)
			if hazard {
				log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "new data hazard: reason=%s", reason)
				u.blockedDataHazard++
				return
			}

			u.pushRunner(ctx, cycle, pending)
			u.pendings.Remove(elem)
			pushedRunners = append(pushedRunners, pending.Runner)
			remaining--
			pushed++
		} else {
			log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "data hazard: reason=%s", reason)
			u.blockedDataHazard++
			return
		}
	}

	for remaining > 0 && !u.pendings.IsFull() {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}
		if pushed > 0 && runner.Runner.InstructionType().IsBranch() {
			u.pendings.Push(runner)
			u.blockedBranch++
			return
		}

		hazard, reason := ctx.IsDataHazard(runner.Runner)
		if !hazard {
			hazard, reason := isHazardWithPushedRunners(pushedRunners, runner.Runner)
			if hazard {
				u.pendings.Push(runner)
				log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "new data hazard: reason=%s", reason)
				u.blockedDataHazard++
				return
			}

			u.pushRunner(ctx, cycle, runner)
			pushedRunners = append(pushedRunners, runner.Runner)
			remaining--
			pushed++
		} else {
			u.pendings.Push(runner)
			log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%s", reason)
			u.blockedDataHazard++
			return
		}
	}
}

func isHazardWithPushedRunners(pushedRunners []risc.InstructionRunner, runner risc.InstructionRunner) (bool, string) {
	pendingWriteRegisters := make(map[risc.RegisterType]int)
	for _, runner := range pushedRunners {
		for _, register := range runner.WriteRegisters() {
			if register == risc.Zero {
				continue
			}
			pendingWriteRegisters[register]++
		}
	}

	for _, register := range runner.ReadRegisters() {
		if register == risc.Zero {
			continue
		}
		if v, exists := pendingWriteRegisters[register]; exists && v > 0 {
			// An instruction needs to read from a register that was updated
			return true, fmt.Sprintf("Read hazard on %s", register)
		}
	}
	return false, ""
}

func (u *controlUnit) pushRunner(ctx *risc.Context, cycle int, runner risc.InstructionRunnerPc) {
	u.outBus.Add(runner, cycle)
	log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner")
}

func (u *controlUnit) flush() {
	u.pendings = comp.NewQueue[risc.InstructionRunnerPc](pendingLength)
}

func (u *controlUnit) isEmpty() bool {
	return u.pendings.Length() == 0
}
