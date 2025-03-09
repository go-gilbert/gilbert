package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/pkg/uflag"
)

const (
	flagWorkDir   = "cwd"
	flagLogFormat = "log-format"
	flagLogLevel  = "log-level"
	flagNoCache   = "no-cache"
)

var workflowFileNames = []string{
	"gilbert.yml",
	"gilbert.yaml",
}

type BootstrapOpts struct {
	WorkDir  string
	LogLevel log.Level
	NoCache  bool
	NoColor  bool
	JSON     bool
}

func (opts BootstrapOpts) buildLogWriter() log.Writer {
	if opts.JSON {
		return log.NewJSONWriter()
	}

	return log.NewConsoleWriter(opts.NoColor)
}

func (opts BootstrapOpts) setupWorkDir() (string, error) {
	workDir := opts.WorkDir
	if workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			err = fmt.Errorf("cannot get current working directory: %w", err)
		}

		return cwd, err
	}

	if err := os.Chdir(workDir); err != nil {
		return "", fmt.Errorf("cannot change working directory: %w", err)
	}

	return workDir, nil
}

func (opts BootstrapOpts) WithDefaults() BootstrapOpts {
	if opts.LogLevel == log.LevelUnknown {
		opts.LogLevel = log.LevelInfo
	}

	return opts
}

type coreFlagsConsumer struct {
	dst *BootstrapOpts
}

func (c coreFlagsConsumer) IsBoolFlag(flagName string) bool {
	// atm all flags require a value.
	return flagName == flagNoCache
}

func (c coreFlagsConsumer) IsKnownFlag(flagName string) bool {
	switch flagName {
	case flagWorkDir, flagLogFormat, flagLogLevel, flagNoCache:
		return true
	}

	return false
}

func (c coreFlagsConsumer) ConsumeFlag(flagName, value string) {
	if flagName == flagNoCache {
		uflag.WriteBoolFlag(&c.dst.NoCache, value)
		return
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	switch flagName {
	case flagWorkDir:
		c.dst.WorkDir = value
	case flagLogFormat:
		c.setLogFormat(value)
	}
}

func (c coreFlagsConsumer) setLogFormat(value string) {
	switch value {
	case log.FormatJSON:
		c.dst.JSON = true
	case log.FormatNoColor:
		c.dst.NoColor = true
	}
}

func (c coreFlagsConsumer) setLogLevel(value string) {
	level := log.ParseLevel(value)
	if level == log.LevelUnknown {
		level = log.LevelInfo
	}

	c.dst.LogLevel = level
}

// BootstrapArgsFromFlags parses command-line flags necessary to configure logging, working directory
// and other core functionality before starting building command-line application from workflow file.
//
// Cobra cannot be used for reading those flags because:
//   - Cobra needs to know all allowed flags, including ones defined in workflow file.
//   - Workflow file path is based on current work dir defined by a flag.
//   - Logger should be configured before cobra to print errors.
func BootstrapArgsFromFlags(args []string) BootstrapOpts {
	var opts BootstrapOpts

	uflag.Parse(coreFlagsConsumer{
		dst: &opts,
	}, args)

	return opts.WithDefaults()
}

func locateWorkflowFile(workspaceDir string) (string, error) {
	for _, fname := range workflowFileNames {
		fPath := filepath.Join(workspaceDir, fname)
		if _, err := os.Stat(fPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return "", fmt.Errorf("cannot access workflow file %q: %w", fPath, err)
		}

		return fPath, nil
	}

	return "", fs.ErrNotExist
}
