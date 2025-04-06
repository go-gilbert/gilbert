package expr

type Option = func(cfg *parseConfig)

// WithDocumentInfo adds information about where expression starts.
//
// Used to specify start line and column position inside a file and provide correct diagnostics.
func WithDocumentInfo(docInfo DocumentInfo) Option {
	return func(cfg *parseConfig) {
		cfg.docInfo = docInfo
	}
}

type parseConfig struct {
	docInfo DocumentInfo
}
