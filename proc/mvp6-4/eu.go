package mvp6_4

import (
	"sort"

	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type euReq struct {
	cycle         int
	app           risc.Application
	invalidations map[int32]bool
}

type euResp struct {
	flush      bool
	sequenceID int32
	pc         int32
	isReturn   bool
	err        error

	invalidation     bool
	invalidationAddr int32
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

	// TODO xxx
	// 4 3 18 | 2 1  -- 29
	// 4 3 18 | 18 1 -- 30
	// 4 3 2  | 18 1 -- 31
	// 4 3 2  | 18 18 -- 32 (theory)
	// 4 3 2  | 18 2  -- 32 (practice, wrong)
	// Problem is instruction 20 that loads into t4 2 instead of 18 (delta=29)
	// 2 read request 64
	// u.runner.Pc == 20 && addrs[0] == 64 && delta == 29
	// Problem: read from memory 2, someone is writing to 64
	//
	// New problem
	// Delta 33
	// Write addr 64
	// 4 3 2 | 1 18 (theory)
	// 4 3 2 | 1 1 (practice)
	//           -- wrong, should stay 18
	addrs := u.runner.Runner.MemoryRead(u.ctx)
	if len(addrs) != 0 {
		return u.ExecuteWithCheckpoint(r, func(r euReq) euResp {
			resp := u.cc.Cycle(ccReq{addrs})
			if !resp.done {
				return euResp{}
			}
			u.memory = resp.memory
			return u.ExecuteWithReset(r, u.run)
		})
	}
	return u.ExecuteWithReset(r, u.run)
}

func (u *executeUnit) run(r euReq) euResp {
	// TODO SW 9
	if u.runner.Pc == -1 {
		//fmt.Println(r.cycle)
	}
	// TODO LW 5
	if u.runner.Pc == -1 {
		//fmt.Println(r.cycle)
	}
	execution, err := u.runner.Runner.Run(u.ctx, r.app.Labels, u.runner.Pc, u.memory)
	if err != nil {
		return euResp{err: err}
	}
	log.Infoi(u.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "execution result: %+v", execution)
	if execution.Return {
		return euResp{isReturn: true}
	}

	if execution.MemoryChange {
		writeAddrs, _ := executionToMemoryChanges(execution)
		// TODO Pending is because we are going to write to the line but we may need to read first from memory
		// TODO Pending represent an intention
		u.execution = execution
		if u.cc.isAddressInL1(writeAddrs) {
			return u.ExecuteWithReset(r, u.memoryChange)
		} else {
			// We need first to fetch the instruction from memory
			return u.ExecuteWithCheckpoint(r, func(r euReq) euResp {
				resp := u.cc.Cycle(ccReq{writeAddrs})
				if !resp.done {
					return euResp{}
				}
				return u.ExecuteWithReset(r, u.memoryChange)
			})
		}
	}

	// TODO We shouldn't write in memory if the cache line is present in another core
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

func (u *executeUnit) memoryChange(r euReq) euResp {
	addrs, memory := executionToMemoryChanges(u.execution)

	alignedAddr, listener, invalidation, err := u.cc.msiWrite(addrs, r.invalidations)
	u.ctx.AddPendingWriteMemoryIntention(getAlignedMemoryAddress(addrs), u.cc.id)
	if err == nil {
		var res euResp
		if invalidation == nil {
			res = euResp{}
		} else {
			res = euResp{
				invalidation:     true,
				invalidationAddr: invalidation.addr,
			}
		}

		if listener == nil {
			// Core is this only owner of the line
			u.cc.writeToL1(addrs, memory)
			u.ctx.DeletePendingRegisters(u.runner.Runner.ReadRegisters(), u.runner.Runner.WriteRegisters())
			u.ctx.DeletePendingWriteMemoryIntention(getAlignedMemoryAddress(addrs), u.cc.id)
			return res
		}

		// In this case, we may have to wait for a write-back
		u.Checkpoint(func(r euReq) euResp {
			var instructions []instructionForNextInvalidate
			pendingWrite := false
		loop:
			for {
				select {
				// TODO Close?
				case resp := <-listener.Ch():
					pendingWrite = pendingWrite || resp.pendingWrite
					if resp.memoryChange != nil {
						instructions = append(instructions, *resp.memoryChange)
					}
				default:
					break loop
				}
			}

			if pendingWrite {
				// We have to wait
				return euResp{}
			}
			u.Reset()

			// TODO Delay to write to L1
			u.cc.writeToL1(addrs, memory)
			for _, instruction := range instructions {
				// Write-back other core changes
				u.cc.writeToL1(instruction.addrs, instruction.memory)
			}
			u.ctx.DeletePendingRegisters(u.runner.Runner.ReadRegisters(), u.runner.Runner.WriteRegisters())
			u.ctx.DeletePendingWriteMemoryIntention(getAlignedMemoryAddress(addrs), u.cc.id)
			return euResp{}
		})

		return res
	} else {
		// The core must write an update on a cache line that was already modified
		// within the same cycle by another core.
		// TODO???
		u.cc.SetInstructionForInvalidateRequest(alignedAddr, addrs, memory)
		//delete(u.ctx.PendingWriteMemoryIntention, getAlignedMemoryAddress(addrs))
		return euResp{}
	}
}

func (u *executeUnit) flush() {
	u.Reset()
	u.sequenceID = 0
}

func (u *executeUnit) isEmpty() bool {
	return u.IsStart()
}
