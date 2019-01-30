package shell

const (
	shellSh         = "/bin/sh"
	shCommandPrefix = "-c"
)

func defaultParams() Params {
	return Params{
		Shell:          shellSh,
		ShellExecParam: shCommandPrefix,
	}
}

func (p *Params) preparedCommand() string {
	return p.Command
}
