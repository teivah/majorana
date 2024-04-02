package mvm5

import (
	"fmt"

	"github.com/teivah/majorana/risc"
)

func logi(ctx *risc.Context, unit string, insType risc.InstructionType, pc int32, detail string, args ...any) {
	if !ctx.Debug {
		return
	}
	fmt.Printf("\t%s: %s (pc=%d, ins=%s)\n", unit, fmt.Sprintf(detail, args...), pc/4, insType)
}

func logu(ctx *risc.Context, unit string, detail string, args ...any) {
	if !ctx.Debug {
		return
	}
	fmt.Printf("\t%s: %s\n", unit, fmt.Sprintf(detail, args...))
}

func log(ctx *risc.Context, detail string, args ...any) {
	if !ctx.Debug {
		return
	}
	fmt.Printf("%s\n", fmt.Sprintf(detail, args...))
}