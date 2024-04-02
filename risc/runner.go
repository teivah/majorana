package risc

type Runner struct {
	Ctx *Context
	App Application
}

func NewRunner(app Application, memoryBytes int) *Runner {
	return &Runner{
		Ctx: NewContext(false, memoryBytes),
		App: app,
	}
}

func (r *Runner) Run() error {
	var pc int32
	for pc/4 < int32(len(r.App.Instructions)) {
		runner := r.App.Instructions[pc/4]
		exe, err := runner.Run(r.Ctx, r.App.Labels, pc)
		if err != nil {
			return err
		}
		if exe.RegisterChange {
			r.Ctx.WriteRegister(exe)
		} else if exe.MemoryChange {
			r.Ctx.WriteMemory(exe)
		}

		if exe.PcChange {
			pc = exe.NextPc
		} else {
			pc += 4
		}
	}
	return nil
}
