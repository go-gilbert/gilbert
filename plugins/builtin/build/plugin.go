package build

import (
	"fmt"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"
	"strings"
)

// Plugin represents Gilbert's plugin
type Plugin struct {
	context *scope.Context
	params  Params
	log     logging.Logger
}

// Call calls a plugin
func (p *Plugin) Call() error {
	cmd, err := p.params.newCompilerProcess(p.context)
	if err != nil {
		return err
	}

	p.log.Debug("Command: '%s'", strings.Join(cmd.Args, " "))
	cmd.Stdout = p.log
	cmd.Stderr = p.log.ErrorWriter()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf(`failed build project, %s`, err)
	}

	if err = cmd.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	return nil
}