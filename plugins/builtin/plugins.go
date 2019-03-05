package builtin

import (
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/plugins/builtin/build"
	"github.com/x1unix/gilbert/plugins/builtin/goget"
	"github.com/x1unix/gilbert/plugins/builtin/shell"
)

// DefaultPlugins is list of default plugins
var DefaultPlugins = map[string]plugins.PluginFactory{
	"build":  build.NewBuildPlugin,
	"shell":  shell.NewShellPlugin,
	"go-get": goget.NewPlugin,
}
