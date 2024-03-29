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
	for r.Ctx.Pc/4 < int32(len(r.App.Instructions)) {
		runner := r.App.Instructions[r.Ctx.Pc/4]
		exe, err := runner.Run(r.Ctx, r.App.Labels)
		if err != nil {
			return err
		}
		r.Ctx.Write(exe)
		r.Ctx.Pc = exe.Pc
	}
	return nil
}
