package expr

import "github.com/go-gilbert/gilbert/pkg/parsetypes"

type DocumentPosition struct {
	// FileName is source file name where expression is located.
	FileName string

	// Position is line and column number position of expression start in a document.
	Position parsetypes.Position

	// ByteOffset is position from start of a document where expression starts.
	ByteOffset int
}

func newDocPosition() DocumentPosition {
	return DocumentPosition{
		Position: parsetypes.NewEmptyPosition(),
	}
}

type parseConfig struct {
	docPos DocumentPosition
}

type Option = func(*parseConfig)

// WithDocumentPosition adds information about where expression starts.
//
// Used to specify start line and column position inside a file and provide correct diagnostics.
func WithDocumentPosition(pos DocumentPosition) Option {
	return func(cfg *parseConfig) {
		cfg.docPos = pos
	}
}

func configFromOptions(opts []Option) parseConfig {
	cfg := parseConfig{
		docPos: newDocPosition(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}
