package mvp6_4

import (
	"fmt"

	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/common/obs"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct {
	ctx                     *risc.Context
	ret                     bool
	pendingBranchResolution bool
	log                     string
	inBus                   *comp.BufferedBus[int32]
	outBus                  *comp.BufferedBus[risc.InstructionRunnerPc]

	pushed      *obs.Gauge
	pendingRead *obs.Gauge
	blocked     *obs.Gauge
}

func newDecodeUnit(ctx *risc.Context, inBus *comp.BufferedBus[int32], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *decodeUnit {
	return &decodeUnit{
		ctx:         ctx,
		inBus:       inBus,
		outBus:      outBus,
		pushed:      &obs.Gauge{},
		pendingRead: &obs.Gauge{},
		blocked:     &obs.Gauge{},
	}
}

func (u *decodeUnit) cycle(cycle int, app risc.Application) {
	pushed := 0
	defer func() {
		u.pushed.Push(pushed)
	}()
	u.pendingRead.Push(u.inBus.PendingRead())
	if u.inBus.CanGet() {
		u.blocked.Push(1)
	} else {
		u.blocked.Push(0)
	}
	if u.ret {
		return
	}
	if u.pendingBranchResolution {
		log.Infou(u.ctx, "DU", "blocked")
		return
	}

	for {
		if !u.outBus.CanAdd() {
			log.Infou(u.ctx, "DU", "can't add")
		}
		pc, exists := u.inBus.Get()
		if !exists {
			return
		}
		if int(pc)/4 >= len(app.Instructions) {
			return
		}
		runner := app.Instructions[pc/4]
		// Clear forward
		runner.Forward(risc.Forward{})
		log.Infoi(u.ctx, "DU", runner.InstructionType(), pc, "decoding")
		jump := false
		if runner.InstructionType().IsUnconditionalBranch() {
			u.pendingBranchResolution = true
			log.Infoi(u.ctx, "DU", runner.InstructionType(), pc, "pending branch resolution")
			u.log = fmt.Sprintf("%v at %d", runner.InstructionType(), pc/4)
			jump = true
		}
		u.outBus.Add(risc.InstructionRunnerPc{
			Runner:     runner,
			Pc:         pc,
			SequenceID: u.ctx.SequenceID(pc),
		}, cycle)
		pushed++
		if jump {
			return
		}
		if runner.InstructionType() == risc.Ret {
			u.ret = true
		}
	}
}

func (u *decodeUnit) notifyBranchResolved() {
	u.pendingBranchResolution = false
}

func (u *decodeUnit) flush() {
	u.pendingBranchResolution = false
	u.ret = false
}

func (u *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
