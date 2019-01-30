package shell

import "strings"

const (
	shellWin             = "cmd.exe"
	winExecParam         = "/C"
	winCodePageFixPrefix = "chcp 65001 > nil" // Force use UTF-8 to provide correct output to stdout
)

func defaultParams() Params {
	return Params{
		Shell:          shellWin,
		ShellExecParam: winExecParam,
	}
}

func (p *Params) preparedCommand() string {
	if !strings.Contains(strings.ToLower(p.Shell), shellWin) {
		// Remove patch for non standard shells (e.g. WSL)
		return p.Command
	}
	return winCodePageFixPrefix + " && " + p.Command
}
