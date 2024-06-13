package cmd

import (
	"github.com/go-gilbert/gilbert/internal/config"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/pkg/containers"
)

type Factory struct {
	cfg *config.LaunchParams

	logger *containers.OnceCell[log.Logger]
}

func NewFactory(cfg *config.LaunchParams) Factory {
	return Factory{
		cfg: cfg,
		logger: containers.NewOnceCell(func() log.Logger {
			level := log.LevelInfo
			if cfg.Debug {
				level = log.LevelDebug
			}

			return log.NewConsoleLogger(level, cfg.NoColor)
		}),
	}
}

func (f Factory) Logger() log.Logger {
	return f.logger.Get()
}

func (f Factory) LaunchParams() *config.LaunchParams {
	return f.cfg
}
