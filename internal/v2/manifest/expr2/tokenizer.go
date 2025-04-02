package expr2

import "github.com/go-gilbert/gilbert/pkg/parsetypes"

type DocumentInfo struct {
	// FileName is document filename to which expression belongs.
	FileName string

	// ByteOffset is offset in bytes where expression starts.
	//
	// Offset is added to expression nodes location information.
	ByteOffset uint

	// StartPosition is line and column number of where expression starts.
	//
	// Added to expression nodes location information.
	StartPosition parsetypes.Position
}

type Option = func(cfg *parserConfig)

type parserConfig struct {
	docInfo DocumentInfo
}

func newParserConfig(opts ...Option) parserConfig {
	cfg := parserConfig{
		docInfo: DocumentInfo{
			FileName:      "",
			ByteOffset:    0,
			StartPosition: parsetypes.NewEmptyPosition(),
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

type parser struct {
	docInfo      DocumentInfo
	prevToken    *token
	currentToken *token
	offset       int
}

func (p *parser) nextToken() token {
	currentPos :=
}
