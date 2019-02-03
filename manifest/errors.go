package manifest

import "fmt"

// PluginConfigError is plugin configuration error
type PluginConfigError struct {
	PluginName    string
	OriginalError error
}

// Error satisfies error interface
func (p *PluginConfigError) Error() string {
	return fmt.Sprintf("failed to read params for plugin '%s', %s", p.PluginName, p.OriginalError)
}

// NewPluginConfigError creates a new PluginConfigError
func NewPluginConfigError(pluginName string, originalError error) *PluginConfigError {
	return &PluginConfigError{
		PluginName:    pluginName,
		OriginalError: originalError,
	}
}
