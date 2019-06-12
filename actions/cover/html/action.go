package html

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/support/shell"

	sdk "github.com/go-gilbert/gilbert-sdk"
)

const (
	coverFilePattern   = "gbcover*.out"
	defaultCoverTarget = "./..."
	defaultTimeout     = sdk.Period(300)
)

// NewAction creates a new html coverage report action handler
func NewAction(scope sdk.ScopeAccessor, params sdk.ActionParams) (h sdk.ActionHandler, err error) {
	handler := &reportAction{alive: true, scope: scope, Timeout: defaultTimeout}
	if err := params.Unmarshal(&handler); err != nil {
		return nil, err
	}

	if len(handler.Packages) == 0 {
		handler.Packages = []string{defaultCoverTarget}
	}

	handler.coverFile, err = ioutil.TempFile(os.TempDir(), coverFilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create coverage temporary file: %s", err)
	}

	return handler, nil
}

type reportAction struct {
	Packages  []string   `mapstructure:"packages"`
	Timeout   sdk.Period `mapstructure:"timeout"`
	scope     sdk.ScopeAccessor
	coverFile *os.File
	alive     bool
}

// Call implements sdk.ActionHandler
func (a *reportAction) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) (err error) {
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

func (a *reportAction) openReport(ctx sdk.JobContextAccessor) error {
	// go tool cover -html=/tmp/cover.out
	cmd := exec.CommandContext(ctx.Context(), "go", "tool", "cover", "-html="+a.coverFile.Name())
	cmd.Dir = a.scope.Environment().ProjectDirectory
	cmd.Stdout = ctx.Log()
	cmd.Stderr = ctx.Log().ErrorWriter()
	ctx.Log().Debugf("cover:html: exec '%s'", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

func (a *reportAction) createReport(ctx sdk.JobContextAccessor) error {
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

func (a *reportAction) clean(ctx sdk.JobContextAccessor) {
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

// Cancel implements sdk.ActionHandler
func (a *reportAction) Cancel(ctx sdk.JobContextAccessor) error {
	a.clean(ctx)
	return nil
}
