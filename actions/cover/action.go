package cover

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert-sdk"

	"github.com/go-gilbert/gilbert/actions/cover/profile"
	"github.com/go-gilbert/gilbert/support/shell"
)

type Action struct {
	scope     sdk.ScopeAccessor
	params    params
	coverFile *os.File
	alive     bool
}

func (a *Action) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) (err error) {
	defer a.clean(ctx)
	cmd, err := a.createCoverCommand(ctx)
	if err != nil {
		return err
	}

	ctx.Log().Debugf("cover command: '%s'", strings.Join(cmd.Args, " "))
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start cover tool, %s", err)
	}

	if err = cmd.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	if !a.alive {
		return nil
	}

	pkgs, err := profile.ConvertProfiles(a.coverFile.Name())
	if err != nil {
		return fmt.Errorf("failed to parse cover profile file, %s", err)
	}

	// TODO: find a better approach to stop on cancel
	if !a.alive {
		return nil
	}

	// Check coverage
	report := profile.Create(*pkgs)
	if err := report.CheckCoverage(a.params.Threshold); err != nil {
		a.printReport(ctx, &report)
		return err
	}

	if a.params.Report {
		a.printReport(ctx, &report)
	}

	return nil
}

func (a *Action) printReport(ctx sdk.JobContextAccessor, r *profile.Report) {
	if r.Total <= 0 {
		ctx.Log().Warnf("No test files found in packages")
	} else {
		prop, desc := a.params.Sort.By, a.params.Sort.Desc

		// Print coverage report only if any data present
		ctx.Log().Info("Coverage report:")
		var str string
		if a.params.FullReport {
			str = r.FormatFull(prop, desc)
		} else {
			str = r.FormatSimple(prop, desc)
		}

		_, _ = ctx.Log().Write([]byte(str))
	}

	ctx.Log().Infof("Total coverage: %.2f%%", r.Percentage())
}

func (a *Action) clean(ctx sdk.JobContextAccessor) {
	if !a.alive {
		return
	}

	a.alive = false
	fname := a.coverFile.Name()
	if err := os.Remove(fname); err != nil {
		ctx.Log().Debugf("failed to remove cover file '%s': %s", fname, err)
		return
	}

	ctx.Log().Debugf("removed cover file '%s'", fname)
}

func (a *Action) createCoverCommand(ctx sdk.JobContextAccessor) (*exec.Cmd, error) {
	// pass package names as is, since '-coverpkg' doesn't recognise them in CSV format (go 1.11+)
	args := make([]string, 0, len(a.params.Packages)+toolArgsPrefixSize)
	args = append(args, "test", "-coverprofile="+a.coverFile.Name())

	for _, pkg := range a.params.Packages {
		val, err := a.scope.ExpandVariables(pkg)
		if err != nil {
			return nil, err
		}

		args = append(args, val)
	}

	cmd := exec.CommandContext(ctx.Context(), "go", args...)
	cmd.Dir = a.scope.Environment().ProjectDirectory
	cmd.Stderr = ctx.Log().ErrorWriter()
	return cmd, nil
}

func (a *Action) Cancel(ctx sdk.JobContextAccessor) error {
	a.clean(ctx)
	return nil
}
