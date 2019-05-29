package cover

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert/support/shell"

	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/actions/cover/report"

	"github.com/go-gilbert/gilbert/actions/cover/profile"
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

	ctx.Log().Debugf("cover: exec '%s'", strings.Join(cmd.Args, " "))

	// "go test" tool sometimes reports errors not to stderr, but to stdout
	// so we also should capture output from stdout
	repFmt := report.NewReportFormatter()
	cmd.Stdout = repFmt
	cmd.Stderr = ctx.Log().ErrorWriter()
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start 'go test' tool, %s", err)
	}

	if err = cmd.Wait(); err != nil {
		a.printFailedPackages(ctx.Log(), repFmt)
		return fmt.Errorf("test execution failed (%s)", shell.FormatExitError(err))
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
	a.printUncoveredItems(ctx.Log(), repFmt)
	prof := profile.Create(*pkgs)
	err = prof.CheckCoverage(a.params.Threshold)
	if err != nil || a.params.Report {
		a.printReport(ctx, &prof)
		return err
	}

	return err
}

func (a *Action) printUncoveredItems(l sdk.Logger, fpFmt *report.Formatter) {
	uncovered, count := fpFmt.UncoveredPackages()
	if count == 0 {
		l.Debug("cover: no uncovered packages in report")
		return
	}

	if !a.params.ShowUncovered {
		l.Warnf("%d packages don't have tests and therefore were not included in the report.", count)
		return
	}

	l.Warnf("%d packages without tests:", count)
	_, _ = l.Write([]byte(uncovered))
}

func (a *Action) printFailedPackages(l sdk.Logger, fpFmt *report.Formatter) {
	failed, count := fpFmt.FailedTests()
	if count == 0 {
		l.Debug("cover: no failed tests available in report")
		return
	}

	l.Errorf("Failed to check test coverage, test run failed with %d errors.\n", count)
	l.Error("Failed tests:")
	_, _ = l.ErrorWriter().Write([]byte(failed))
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
		ctx.Log().Debugf("cover: failed to remove cover file '%s': %s", fname, err)
		return
	}

	ctx.Log().Debugf("cover: removed cover file '%s'", fname)
}

func (a *Action) createCoverCommand(ctx sdk.JobContextAccessor) (*exec.Cmd, error) {
	// pass package names as is, since '-coverpkg' doesn't recognise them in CSV format (go 1.11+)
	args := make([]string, 0, len(a.params.Packages)+toolArgsPrefixSize)
	args = append(args, "test", "-coverprofile="+a.coverFile.Name(), "-json")

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
