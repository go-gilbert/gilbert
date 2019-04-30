package loader

// LoadLibrary loads library from provided source
func LoadLibrary(libPath string) (sdk.PluginFactory, string, error) {
	return nil, "", errors.New("plugins currently are not supported on Windows")
}
