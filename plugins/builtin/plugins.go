package builtin

import (
	"github.com/x1unix/guru/plugins"
	"github.com/x1unix/guru/plugins/builtin/build"
	"github.com/x1unix/guru/plugins/builtin/shell"
)

// DefaultPlugins is list of default plugins
var DefaultPlugins = map[string]plugins.PluginFactory{
	"build": build.NewBuildPlugin,
	"shell": shell.NewShellPlugin,
}
