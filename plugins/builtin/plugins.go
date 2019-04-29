package builtin

import (
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/plugins/builtin/build"
	"github.com/go-gilbert/gilbert/plugins/builtin/cover"
	"github.com/go-gilbert/gilbert/plugins/builtin/goget"
	"github.com/go-gilbert/gilbert/plugins/builtin/shell"
	"github.com/go-gilbert/gilbert/plugins/builtin/watch"
)

// DefaultPlugins is list of default plugins
var DefaultPlugins = map[string]sdk.PluginFactory{
	"build":  build.NewBuildPlugin,
	"shell":  shell.NewShellPlugin,
	"go-get": goget.NewPlugin,
	"watch":  watch.NewWatchPlugin,
	"cover":  cover.NewPlugin,
}
