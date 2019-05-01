package support

import (
	"os"
	"runtime"
)

// PluginPermissions is permissions for plugin assets
var PluginPermissions = os.FileMode(0755)

// PluginExtension adds plugin extension format to the provided plugin file
func AddPluginExtension(fileName string) string {
	if runtime.GOOS == "windows" {
		return fileName + ".dll"
	}

	return fileName + ".so"
}
