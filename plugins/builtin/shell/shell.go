package shell

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"
	"os"
	"os/exec"
	"runtime"
	"syscall"
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

func (p *Params) createProcess(ctx *scope.Scope) (*exec.Cmd, error) {
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
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	return cmd, nil
}

func newParams(ctx *scope.Scope) Params {
	p := defaultParams()
	p.WorkDir = ctx.Environment.ProjectDirectory

	return p
}

// NewShellPlugin creates a new shell plugin instance
func NewShellPlugin(scope *scope.Scope, params manifest.RawParams, log logging.Logger) (plugins.Plugin, error) {
	p := newParams(scope)

	if err := mapstructure.Decode(params, &p); err != nil {
		return nil, fmt.Errorf("failed to read configuration: %s", err)
	}

	return &Plugin{
		scope:  scope,
		params: p,
		log:    log,
		done:   make(chan bool, 1),
	}, nil
}
