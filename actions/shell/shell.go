package shell

import (
	"os"
	"os/exec"

	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/tools/shell"
)

// Params contains params for shell plugin
type Params struct {
	// Command is command to execute
	Command string

	// Silent param hides stdout and stderr from output
	Silent bool

	// RawOutput removes logging output decoration from stdout and stderr
	RawOutput bool

	// Shell is default shell to start
	Shell string

	// ShellExecParam is param used by shell to pass command.
	//
	// Example: "bash -c "your command"
	ShellExecParam string

	// WorkDir is current working directory
	WorkDir string

	// Env is set of environment variables
	Env shell.Environment
}

func (p *Params) createProcess(ctx sdk.ScopeAccessor) (*exec.Cmd, error) {
	// TODO: check if Shell or ShellExecParam are empty
	cmdstr, err := ctx.ExpandVariables(p.preparedCommand())
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(p.Shell, p.ShellExecParam, cmdstr)

	wd, err := ctx.ExpandVariables(p.WorkDir)
	if err != nil {
		return nil, err
	}

	cmd.Dir = wd

	// TODO: inherit global vars
	if !p.Env.Empty() {
		cmd.Env = p.Env.ToArray(os.Environ()...)
	} else {
		cmd.Env = os.Environ()
	}

	// Assign process group (for unix only)
	decorateCommand(cmd)

	return cmd, nil
}

func newParams(ctx sdk.ScopeAccessor) Params {
	p := defaultParams()
	p.WorkDir = ctx.Environment().ProjectDirectory

	return p
}

// NewAction creates a new shell action handler instance
func NewAction(scope sdk.ScopeAccessor, params sdk.ActionParams) (sdk.ActionHandler, error) {
	p := newParams(scope)

	if err := params.Unmarshal(&p); err != nil {
		return nil, err
	}

	return &Action{
		scope:  scope,
		params: p,
	}, nil
}
