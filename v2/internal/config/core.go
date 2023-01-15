package config

import (
	"fmt"
	"github.com/go-gilbert/gilbert/v2/internal/spec"
	"os"
	"path/filepath"

	"github.com/go-gilbert/gilbert/v2/internal/log"
)

const (
	DefaultSpecFile  = "gilbert.hcl"
	DefaultLogFormat = "color"
)

type CoreConfig struct {
	WorkDir   string
	SpecFile  string
	LogFormat string
	Verbose   bool
}

func (cfg *CoreConfig) NewLogger() (*log.OutputPrinter, error) {
	level := log.InfoLevel
	if cfg.Verbose {
		level = log.DebugLevel
	}

	writer, err := log.EncoderFromString(cfg.LogFormat)
	if err != nil {
		return nil, err
	}

	return log.NewOutputPrinter(level, log.NewStdoutWriter(), writer), nil
}

func (cfg *CoreConfig) ProjectSpec() spec.ProjectSpec {
	fileName := cfg.SpecFile
	if !filepath.IsAbs(cfg.SpecFile) {
		fileName = filepath.Join(cfg.WorkDir, cfg.SpecFile)
	}

	return spec.ProjectSpec{
		FileName:         fileName,
		WorkingDirectory: cfg.WorkDir,
	}
}

type Config struct {
	CoreConfig
	CacheDir string
}

// NewCoreConfig returns new CoreConfig with current working directory.
func NewCoreConfig() (*CoreConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	cfg := &CoreConfig{
		WorkDir:   cwd,
		SpecFile:  DefaultSpecFile,
		LogFormat: DefaultLogFormat,
	}
	return cfg, nil
}
