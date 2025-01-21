package html

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/runner"
	"github.com/go-gilbert/gilbert/internal/runner/job"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/internal/support/shell"
)

const (
	coverFilePattern   = "gbcover*.out"
	defaultCoverTarget = "./..."
	defaultTimeout     = manifest.Period(300)
)

// NewAction creates a new html coverage report action handler
func NewAction(scope *scope.Scope, params manifest.ActionParams) (h runner.ActionHandler, err error) {
	handler := &reportAction{alive: true, scope: scope, Timeout: defaultTimeout}
	if err := params.Unmarshal(&handler); err != nil {
		return nil, err
	}

	if len(handler.Packages) == 0 {
		handler.Packages = []string{defaultCoverTarget}
	}

	handler.coverFile, err = os.CreateTemp(os.TempDir(), coverFilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create coverage temporary file: %w", err)
	}

	return handler, nil
}

type reportAction struct {
	Packages  []string        `mapstructure:"packages"`
	Timeout   manifest.Period `mapstructure:"timeout"`
	scope     *scope.Scope
	coverFile *os.File
	alive     bool
}

func (a *reportAction) Call(ctx *job.RunContext, r *runner.TaskRunner) (err error) {
	defer a.clean(ctx)
	ctx.Log().Info("Generating coverage profile...")
	if err := a.createReport(ctx); err != nil {
		return err
	}

	if !a.alive {
		return nil
	}

	ctx.Log().Info("Opening report...")
	if err := a.openReport(ctx); err != nil {
		return fmt.Errorf("failed to open report in browser, %s", err)
	}

	time.Sleep(a.Timeout.ToDuration())
	return nil
}

func (a *reportAction) openReport(ctx *job.RunContext) error {
	// go tool cover -html=/tmp/cover.out
	cmd := exec.CommandContext(ctx.Context(), "go", "tool", "cover", "-html="+a.coverFile.Name())
	cmd.Dir = a.scope.Environment().ProjectDirectory
	cmd.Stdout = ctx.Log()
	cmd.Stderr = ctx.Log().ErrorWriter()
	ctx.Log().Debugf("cover:html: exec '%s'", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

func (a *reportAction) createReport(ctx *job.RunContext) error {
	// pass package names as is, since '-coverpkg' doesn't recognise them in CSV format (go 1.11+)
	args := []string{"test", "-coverprofile=" + a.coverFile.Name()}
	for _, pkg := range a.Packages {
		val, err := a.scope.ExpandVariables(pkg)
		if err != nil {
			return err
		}

		args = append(args, val)
	}

	//cmd := exec.CommandContext(ctx.Context(), "go", args...)
	cmd := exec.CommandContext(ctx.Context(), "go", args...)
	cmd.Dir = a.scope.Environment().ProjectDirectory
	cmd.Stdout = ctx.Log()
	cmd.Stderr = ctx.Log().ErrorWriter()

	ctx.Log().Debugf("cover:html: exec '%s'", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test execution failed (%s)", shell.FormatExitError(err))
	}

	return nil
}

func (a *reportAction) clean(ctx *job.RunContext) {
	if !a.alive {
		return
	}

	a.alive = false
	fname := a.coverFile.Name()
	if err := os.Remove(fname); err != nil {
		ctx.Log().Debugf("cover:html: failed to remove cover file '%s': %s", fname, err)
		return
	}

	ctx.Log().Debugf("cover:html: removed cover file '%s'", fname)
}

func (a *reportAction) Cancel(ctx *job.RunContext) error {
	a.clean(ctx)
	return nil
}
