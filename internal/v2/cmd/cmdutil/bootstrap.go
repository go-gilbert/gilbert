package cmdutil

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
	FlagWorkDir   = "cwd"
	FlagLogFormat = "log-format"
	FlagLogLevel  = "log-level"
	FlagNoCache   = "no-cache"
)

const DefaultWorkflowFilename = "gilbert.yaml"

var workflowFileNames = []string{
	DefaultWorkflowFilename,
	"gilbert.yml",
}

type BootstrapOpts struct {
	WorkDir  string
	LogLevel log.Level
	NoCache  bool
	NoColor  bool
	JSON     bool
}

func (opts BootstrapOpts) BuildLogWriter() log.Writer {
	if opts.JSON {
		return log.NewJSONWriter()
	}

	return log.NewConsoleWriter(opts.NoColor)
}

func (opts BootstrapOpts) SetupWorkDir() (string, error) {
	workDir := opts.WorkDir
	if workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			err = fmt.Errorf("cannot get current working directory: %w", err)
		}

		return cwd, err
	}

	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("unable to resolve absolute path for a working directory: %w", err)
	}

	if err := os.Chdir(absWorkDir); err != nil {
		return "", fmt.Errorf("cannot change working directory: %w", err)
	}

	return absWorkDir, nil
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
	return flagName == FlagNoCache
}

func (c coreFlagsConsumer) IsKnownFlag(flagName string) bool {
	switch flagName {
	case FlagWorkDir, FlagLogFormat, FlagLogLevel, FlagNoCache:
		return true
	}

	return false
}

func (c coreFlagsConsumer) ConsumeFlag(flagName, value string) {
	if flagName == FlagNoCache {
		uflag.WriteBoolFlag(&c.dst.NoCache, value)
		return
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	switch flagName {
	case FlagWorkDir:
		c.dst.WorkDir = value
	case FlagLogFormat:
		c.setLogFormat(value)
	case FlagLogLevel:
		c.setLogLevel(value)
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

func LocateWorkflowFile(workspaceDir string) (string, error) {
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
