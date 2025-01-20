// +build linux darwin

package support

/////////////////////////////////////////////////////
// Codebase for platforms that supports Go plugins //
/////////////////////////////////////////////////////

// AddPluginExtension adds plugin extension format to the provided plugin file
func AddPluginExtension(fileName string) string {
	return fileName + ".so"
}

// BuildMode is build mode for plugins
var BuildMode = "plugin"
