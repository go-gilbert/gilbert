package manifest

import "fmt"

type PluginConfigError struct {
	PluginName    string
	OriginalError error
}

func (p *PluginConfigError) Error() string {
	return fmt.Sprintf("failed to read params for plugin '%s', %s", p.PluginName, p.OriginalError)
}

func NewPluginConfigError(pluginName string, originalError error) *PluginConfigError {
	return &PluginConfigError{
		PluginName:    pluginName,
		OriginalError: originalError,
	}
}
