// +build freebsd

package support

// PluginExtension adds plugin extension format to the provided plugin file
func AddPluginExtension(fileName string) string {
	return fileName
}

// BuildMode is build mode for plugins
var BuildMode = "exe"
