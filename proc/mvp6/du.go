package mvp6

import (
	"fmt"

	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct {
	ret                     bool
	pendingBranchResolution bool
	log                     string
	inBus                   *comp.BufferedBus[int32]
	outBus                  *comp.BufferedBus[risc.InstructionRunnerPc]
}

func newDecodeUnit(inBus *comp.BufferedBus[int32], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *decodeUnit {
	return &decodeUnit{inBus: inBus, outBus: outBus}
}

func (u *decodeUnit) cycle(cycle int, app risc.Application, ctx *risc.Context) {
	if u.ret {
		return
	}
	if u.pendingBranchResolution {
		log.Infou(ctx, "DU", "blocked")
		return
	}

	for {
		if !u.outBus.CanAdd() {
			log.Infou(ctx, "DU", "can't add")
		}
		pc, exists := u.inBus.Get()
		if !exists {
			return
		}
		if int(pc)/4 >= len(app.Instructions) {
			return
		}
		runner := app.Instructions[pc/4]
		log.Infoi(ctx, "DU", runner.InstructionType(), pc/4, "decoding")
		jump := false
		if runner.InstructionType().IsUnconditionalBranch() {
			u.pendingBranchResolution = true
			u.log = fmt.Sprintf("%v at %d", runner.InstructionType(), pc/4)
			jump = true
		}
		u.outBus.Add(risc.InstructionRunnerPc{
			Runner: runner,
			Pc:     pc,
		}, cycle)
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
}

func (u *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
