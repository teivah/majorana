package mvp7_1

import (
	"sort"

	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type euReq struct {
	cycle int
	app   risc.Application
}

type euResp struct {
	flush      bool
	sequenceID int32
	pc         int32
	isReturn   bool
	err        error
}

type executeUnit struct {
	ctx *risc.Context
	co.Coroutine[euReq, euResp]
	bu     *btbBranchUnit
	inBus  *comp.BufferedBus[*risc.InstructionRunnerPc]
	outBus *comp.BufferedBus[risc.ExecutionContext]
	mmu    *memoryManagementUnit
	cc     *cacheController

	// Pending
	memory     []int8
	runner     risc.InstructionRunnerPc
	sequenceID int32
	execution  risc.Execution
}

func newExecuteUnit(ctx *risc.Context, bu *btbBranchUnit, inBus *comp.BufferedBus[*risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.ExecutionContext], mmu *memoryManagementUnit, cc *cacheController) *executeUnit {
	eu := &executeUnit{
		ctx:    ctx,
		bu:     bu,
		inBus:  inBus,
		outBus: outBus,
		mmu:    mmu,
		cc:     cc,
	}
	eu.Coroutine = co.New(eu.start)
	eu.Coroutine.Pre(func(r euReq) bool {
		if eu.sequenceID == 0 || eu.runner.Runner == nil {
			return false
		}
		if eu.runner.SequenceID > eu.sequenceID {
			eu.flush()
			return true
		}
		return false
	})
	return eu
}

func (u *executeUnit) start(r euReq) euResp {
	runner, exists := u.inBus.Get()
	if !exists {
		return euResp{}
	}
	u.runner = *runner
	return u.ExecuteWithCheckpoint(r, u.prepareRun)
}

func (u *executeUnit) prepareRun(r euReq) euResp {
	if !u.outBus.CanAdd() {
		log.Infou(u.ctx, "EU", "can't add")
		return euResp{}
	}

	if u.runner.Receiver != nil {
		var value int32
		select {
		case v := <-u.runner.Receiver:
			log.Infoi(u.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "receive forward register value %d", v)
			value = v
		default:
			return euResp{}
		}

		u.runner.Runner.Forward(risc.Forward{Value: value, Register: u.runner.ForwardRegister})
		u.runner.Receiver = nil
	}

	// Create the branch unit assertions
	u.bu.assert(u.runner)

	log.Infoi(u.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "executing")

	addrs := u.runner.Runner.MemoryRead(u.ctx)
	if len(addrs) != 0 {
		return u.ExecuteWithCheckpoint(r, func(r euReq) euResp {
			resp := u.cc.read.Cycle(ccReadReq{r.cycle, addrs})
			if !resp.done {
				return euResp{}
			}
			u.memory = resp.data
			return u.ExecuteWithReset(r, u.run)
		})
	}
	return u.ExecuteWithReset(r, u.run)
}

func (u *executeUnit) run(r euReq) euResp {
	execution, err := u.runner.Runner.Run(u.ctx, r.app.Labels, u.runner.Pc, u.memory)
	if err != nil {
		return euResp{err: err}
	}
	log.Infoi(u.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "execution result: %+v", execution)
	if execution.Return {
		return euResp{isReturn: true}
	}

	if execution.MemoryChange {
		writeAddrs, data := executionToMemoryChanges(execution)
		u.execution = execution

		return u.ExecuteWithCheckpoint(r, func(r euReq) euResp {
			resp := u.cc.write.Cycle(ccWriteReq{r.cycle, writeAddrs, data})
			if resp.done {
				u.Reset()
			}
			return euResp{}
		})
	}

	u.outBus.Add(risc.ExecutionContext{
		SequenceID:      u.runner.SequenceID,
		Execution:       execution,
		InstructionType: u.runner.Runner.InstructionType(),
		WriteRegisters:  u.runner.Runner.WriteRegisters(),
		ReadRegisters:   u.runner.Runner.ReadRegisters(),
	}, r.cycle)

	if u.runner.Forwarder == nil {
		if u.runner.Runner.InstructionType().IsUnconditionalBranch() {
			log.Infoi(u.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc,
				"notify jump address resolved from %d to %d", u.runner.Pc/4, execution.NextPc/4)
			u.bu.notifyUnconditionalJumpAddressResolved(u.runner.Pc, execution.NextPc)
		}
		if u.runner.Runner.InstructionType().IsConditionalBranch() {
			if execution.PcChange {
				// Branch taken (jump)
				u.bu.notifyConditionalBranchTaken(u.runner.SequenceID)
			} else {
				// Branch not taken (next PC)
				u.bu.notifyConditionalBranchNotTaken()
			}
		}
		if execution.PcChange && u.bu.shouldFlushPipeline(execution.NextPc) {
			log.Infoi(u.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "should be a flush")
			return euResp{flush: true, sequenceID: u.runner.SequenceID, pc: execution.NextPc}
		}
	} else {
		u.runner.Forwarder <- execution.RegisterValue
		log.Infoi(u.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "forward register value %d", execution.RegisterValue)
		if u.runner.Runner.InstructionType().IsBranch() {
			panic("shouldn't be a branch")
		}
	}

	return euResp{}
}

func executionToMemoryChanges(execution risc.Execution) ([]int32, []int8) {
	type change struct {
		addr   int32
		change int8
	}
	var changes []change
	for a, v := range execution.MemoryChanges {
		changes = append(changes, change{
			addr:   a,
			change: v,
		})
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].addr < changes[j].addr
	})

	var addrs []int32
	var memory []int8

	for _, c := range changes {
		addrs = append(addrs, c.addr)
		memory = append(memory, c.change)
	}

	return addrs, memory
}

func (u *executeUnit) flush() {
	u.Reset()
	u.sequenceID = 0
	u.cc.flush()
}

func (u *executeUnit) isEmpty() bool {
	return u.IsStart()
}
