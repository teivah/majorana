package log

import (
	"fmt"

	"github.com/teivah/majorana/risc"
)

func Infoi(ctx *risc.Context, unit string, insType risc.InstructionType, pc int32, detail string, args ...any) {
	if !ctx.Debug {
		return
	}
	fmt.Printf("\t%s: %s (pc=%d, ins=%s)\n", unit, fmt.Sprintf(detail, args...), pc/4, insType)
}

func Infou(ctx *risc.Context, unit string, detail string, args ...any) {
	if !ctx.Debug {
		return
	}
	fmt.Printf("\t%s: %s\n", unit, fmt.Sprintf(detail, args...))
}

func Info(ctx *risc.Context, detail string, args ...any) {
	if !ctx.Debug {
		return
	}
	fmt.Printf("%s\n", fmt.Sprintf(detail, args...))
}
