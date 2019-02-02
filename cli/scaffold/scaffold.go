package scaffold

import (
	"github.com/urfave/cli"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
)

func ScaffoldManifest(c *cli.Context) (err error) {
	logging.Log.Debug("debugging is %v", scope.Debug)
	return nil
}
