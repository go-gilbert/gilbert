package maintenance

import (
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/storage"
	"github.com/urfave/cli"
)

const (
	targetAll     = "all"
	targetPlugins = "plugins"
)

var (
	ClearAllFlag = cli.BoolFlag{
		Name:  targetAll,
		Usage: "clear everything",
	}

	ClearPluginsFlag = cli.BoolFlag{
		Name:  targetPlugins,
		Usage: "clear downloaded plugins",
	}
)

func ClearCacheAction(ctx *cli.Context) (err error) {
	defer func() {
		if err == nil {
			log.Default.Success("Done!")
		}
	}()

	if ctx.NumFlags() == 0 {
		log.Default.Log("Nothing to clear!")
		return
	}

	if ctx.Bool(targetAll) {
		log.Default.Log("Clearing Gilbert storage...")
		return storage.Delete(storage.Root)
	}

	if ctx.Bool(targetPlugins) {
		log.Default.Log("Clearing downloaded plugins...")
		if err = storage.Delete(storage.Plugins); err != nil {
			return err
		}
	}
	return nil
}
