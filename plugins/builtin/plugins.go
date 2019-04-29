package builtin

import (
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/x1unix/gilbert/plugins/builtin/build"
	"github.com/x1unix/gilbert/plugins/builtin/cover"
	"github.com/x1unix/gilbert/plugins/builtin/goget"
	"github.com/x1unix/gilbert/plugins/builtin/shell"
	"github.com/x1unix/gilbert/plugins/builtin/watch"
)

// DefaultPlugins is list of default plugins
var DefaultPlugins = map[string]sdk.PluginFactory{
	"build":  build.NewBuildPlugin,
	"shell":  shell.NewShellPlugin,
	"go-get": goget.NewPlugin,
	"watch":  watch.NewWatchPlugin,
	"cover":  cover.NewPlugin,
}
