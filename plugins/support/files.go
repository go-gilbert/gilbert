package support

import "runtime"

// PluginExtension returns plugin extension format
func PluginExtension() string {
	if runtime.GOOS == "windows" {
		return "dll"
	}

	return "so"
}
