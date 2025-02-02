package tasks

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"github.com/go-gilbert/gilbert/internal/actions"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/plugins"
	"github.com/go-gilbert/gilbert/internal/runner"
	"github.com/go-gilbert/gilbert/internal/scope"

	"github.com/urfave/cli"
)

const (
	// OverrideVarFlag is flag name for custom variable values
	OverrideVarFlag = "var"

	varDelimiter = "="
	paramsCount  = 2
)

func wrapManifestError(parent error) error {
	return fmt.Errorf("%s\n\nCheck if 'gilbert.yaml' file exists or has correct syntax and check all import statements", parent)
}

// RunTask is a handler for 'run' command
func RunTask(c *cli.Context) (err error) {
	// Read cmd args
	if c.NArg() == 0 {
		return fmt.Errorf("no task specified")
	}

	task := c.Args()[0]

	// Get working dir and read manifest
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current working directory, %v", err)
	}

	man, err := manifest.FromDirectory(cwd)
	if err != nil {
		return wrapManifestError(err)
	}

	// Prepare context and import plugins
	ctx, cancelFn := context.WithCancel(context.Background())

	if err := importProjectPlugins(ctx, man, cwd); err != nil {
		cancelFn()
		return wrapManifestError(err)
	}

	// TODO: inject plugins
	cfg := runner.Config{
		Logger:   log.Default,
		Handlers: runner.NewHandlerSet(actions.BuiltinHandlers),
		Manifest: man,
		WorkDir:  cwd,
	}
	tr := runner.NewTaskRunner(cfg)
	tr.SetContext(ctx, cancelFn)
	go handleShutdown(cancelFn)

	// get variables passed with '--var' flags
	vars := getOverrideVars(c)
	if err := tr.Run(task, vars); err != nil {
		return err
	}

	log.Default.Successf("Task %q completed successfully\n", task)
	return nil
}

func getOverrideVars(c *cli.Context) manifest.Vars {
	ss := c.StringSlice(OverrideVarFlag)
	sLen := len(ss)

	if sLen == 0 {
		return nil
	}

	out := make(manifest.Vars, sLen)
	for _, s := range ss {
		if s == "" {
			continue
		}

		// param=value
		vals := strings.Split(s, varDelimiter)
		if len(vals) < paramsCount {
			continue
		}

		key := strings.TrimSpace(vals[0])
		val := vals[1]
		log.Default.Debugf("cmd: set variable %q = %q", key, val)
		out[key] = val
	}

	return out
}

func importProjectPlugins(ctx context.Context, m *manifest.Manifest, cwd string) error {
	if len(m.Plugins) == 0 {
		return nil
	}

	if runtime.GOOS == "windows" {
		log.Default.Warn("Warning: plugins currently are not supported on this platform")
		return nil
	}

	s := scope.CreateScope(m.Parser, cwd, m.Vars)
	for _, uri := range m.Plugins {
		expanded, err := s.ExpandVariables(uri)
		if err != nil {
			return fmt.Errorf("failed to load plugins from manifest, %s", err)
		}

		if err := plugins.Import(ctx, expanded); err != nil {
			return err
		}
	}

	return nil
}

func handleShutdown(cancelFn context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Default.Log("Shutting down...")
		cancelFn()
	}
}
