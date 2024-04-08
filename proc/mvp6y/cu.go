package mvp6y

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/common/obs"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	pendingLength = 10
)

type controlUnit struct {
	inBus    *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus   *comp.BufferedBus[*risc.InstructionRunnerPc]
	pendings *comp.Queue[risc.InstructionRunnerPc]

	pushed            *obs.Gauge
	pending           *obs.Gauge
	pendingRead       *obs.Gauge
	blocked           *obs.Gauge
	forwarding        int
	total             int
	cantAdd           int
	blockedBranch     int
	blockedDataHazard int
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[*risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:       inBus,
		outBus:      outBus,
		pendings:    comp.NewQueue[risc.InstructionRunnerPc](pendingLength),
		pushed:      &obs.Gauge{},
		pending:     &obs.Gauge{},
		pendingRead: &obs.Gauge{},
		blocked:     &obs.Gauge{},
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	pushed := 0
	defer func() {
		u.pushed.Push(pushed)
		u.pending.Push(u.pendings.Length())
	}()
	u.pendingRead.Push(u.inBus.PendingRead())
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
	var pushedRunners []*risc.InstructionRunnerPc
	skippedRegisterReadForMemoryWrite := make(map[risc.RegisterType]bool)
	skippedRegisterWrite := make(map[risc.RegisterType]bool)

	defer func() {
		for _, runner := range pushedRunners {
			ctx.AddPendingRegisters(runner.Runner)
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
			if hazard, register := isMemoryHazard(skippedRegisterReadForMemoryWrite, nil, pending.Runner); hazard {
				log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "memory hazard on register %s", register)
				u.blockedDataHazard++
				// TODO Why?
				return
			}

			hazard, conflictRunner, register := isHazardWithPushedRunners(pushedRunners, pending.Runner)
			if hazard {
				if pending.Runner.InstructionType().IsMemoryRead() {
					log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "memory hazard on register %s", register)
					return
				}

				ch := make(chan int32, 1)
				conflictRunner.Forwarder = ch
				conflictRunner.ForwardRegister = register
				pending.Receiver = ch
				pending.ForwardRegister = register
				log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "pushing forward runner")
				pushed++
				u.forwarding++
			} else {
				log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "pushing runner")
			}

			u.outBus.Add(&pending, cycle)
			u.pendings.Remove(elem)
			pushedRunners = append(pushedRunners, &pending)
			remaining--
			pushed++

			if pending.Runner.InstructionType().IsBranch() {
				return
			}
		} else {
			log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "data hazard: reason=%s", reason)
			if pending.Runner.InstructionType().IsMemoryWrite() {
				for _, register := range pending.Runner.ReadRegisters() {
					skippedRegisterReadForMemoryWrite[register] = true
				}
			}
			if len(pending.Runner.WriteRegisters()) > 0 {
				for _, register := range pending.Runner.WriteRegisters() {
					skippedRegisterWrite[register] = true
				}
			}
			u.blockedDataHazard++
			// TODO Return?
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
			if hazard, register := isMemoryHazard(skippedRegisterReadForMemoryWrite, skippedRegisterWrite, runner.Runner); hazard {
				log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "memory hazard on register %s", register)
				u.blockedDataHazard++
				u.pendings.Push(runner)
				// TODO Why?
				return
			}

			hazard, conflictRunner, register := isHazardWithPushedRunners(pushedRunners, runner.Runner)
			if hazard {
				if runner.Runner.InstructionType().IsMemoryRead() {
					log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "memory hazard on register %s", register)
					u.pendings.Push(runner)
					return
				}

				ch := make(chan int32, 1)
				conflictRunner.Forwarder = ch
				conflictRunner.ForwardRegister = register
				runner.Receiver = ch
				runner.ForwardRegister = register
				log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing forward runner")
				pushed++
				u.forwarding++
			} else {
				log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner")
			}

			u.outBus.Add(&runner, cycle)
			pushedRunners = append(pushedRunners, &runner)
			remaining--
			pushed++

			if runner.Runner.InstructionType().IsBranch() {
				return
			}
		} else {
			u.pendings.Push(runner)
			if runner.Runner.InstructionType().IsMemoryWrite() {
				for _, register := range runner.Runner.ReadRegisters() {
					skippedRegisterReadForMemoryWrite[register] = true
				}
			}
			if len(runner.Runner.WriteRegisters()) > 0 {
				for _, register := range runner.Runner.WriteRegisters() {
					skippedRegisterWrite[register] = true
				}
			}
			log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%s", reason)
			u.blockedDataHazard++
			if runner.Runner.InstructionType().IsBranch() {
				return
			}
		}
	}
}

func isHazardWithPushedRunners(pushedRunners []*risc.InstructionRunnerPc, runner risc.InstructionRunner) (bool, *risc.InstructionRunnerPc, risc.RegisterType) {
	pendingWriteRegisters := make(map[risc.RegisterType]*risc.InstructionRunnerPc)
	for _, runner := range pushedRunners {
		for _, register := range runner.Runner.WriteRegisters() {
			if register == risc.Zero {
				continue
			}
			pendingWriteRegisters[register] = runner
		}
	}

	for _, register := range runner.ReadRegisters() {
		if register == risc.Zero {
			continue
		}
		if v, exists := pendingWriteRegisters[register]; exists {
			// An instruction needs to read from a register that was updated
			return true, v, register
		}
	}
	return false, nil, risc.Zero
}

func isMemoryHazard(skippedRegisterReadForMemoryWrite map[risc.RegisterType]bool, skippedRegisterWrite map[risc.RegisterType]bool, runner risc.InstructionRunner) (bool, risc.RegisterType) {
	// Prevent such memory hazards:
	//
	// sw t0, 0, zero    # memory[0] = t0
	// addi t0, zero, 42 # t0 = 42
	//
	// If sw is skipped because of a memory hazard, and that addi is executed first,
	// memory[0] will be equal to the wrong value.
	for _, register := range runner.WriteRegisters() {
		if register == risc.Zero {
			continue
		}
		if skippedRegisterReadForMemoryWrite[register] {
			return true, register
		}
	}

	// Prevent such memory hazards:
	//
	// add    t1, t0, a0
	// lb     t1, 0(t1)
	//
	// The second instruction reads from t1, so it has to wait for the first
	// instruction to be written
	for _, register := range runner.ReadRegisters() {
		if register == risc.Zero {
			continue
		}
		if skippedRegisterWrite[register] {
			return true, register
		}
	}
	return false, risc.Zero
}

func (u *controlUnit) flush() {
	u.pendings = comp.NewQueue[risc.InstructionRunnerPc](pendingLength)
}

func (u *controlUnit) isEmpty() bool {
	return u.pendings.Length() == 0
}
