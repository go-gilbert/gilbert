package expr

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/conf"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

// EvalConfig contains settings for expression interpreter
type EvalConfig = conf.Config

type DocumentInfo struct {
	// FileName is document filename to which expression belongs.
	FileName string `json:"fileName"`

	// ByteOffset is Offset in bytes where expression starts.
	//
	// This value affects all Offset numbers returned by Tokenizer and errors.
	ByteOffset int `json:"byteOffset"`

	// StartPosition is line and column number of where expression starts.
	//
	// Added to expression nodes location information.
	//
	// Note: Column value should be one character before start of expression.
	// That means, if string starts at beginning of a string - Column should be 0.
	StartPosition parsetypes.Position `json:"startPosition"`
}

func (docInfo DocumentInfo) newOffsetRange(start, end int) parsetypes.OffsetRange {
	return parsetypes.OffsetRange{
		Start: docInfo.ByteOffset + start,
		End:   docInfo.ByteOffset + end,
	}
}

func (docInfo DocumentInfo) location() *parsetypes.Location {
	return &parsetypes.Location{
		FileName: docInfo.FileName,
		Offset:   parsetypes.NewOffsetRange(docInfo.ByteOffset, docInfo.ByteOffset),
		Range:    parsetypes.NewRange(docInfo.StartPosition, docInfo.StartPosition),
	}
}

type Option = func(cfg *parseConfig)

// WithDocumentInfo adds information about where expression starts.
//
// Used to specify start line and column position inside a file and provide correct diagnostics.
func WithDocumentInfo(docInfo DocumentInfo) Option {
	return func(cfg *parseConfig) {
		cfg.docInfo = docInfo
	}
}

// WithEvalConfig sets custom config for eval expression interpreter.
func WithEvalConfig(opts ...expr.Option) Option {
	return func(cfg *parseConfig) {
		for _, opt := range opts {
			opt(cfg.evalCfg)
		}
	}
}

type parseConfig struct {
	docInfo DocumentInfo
	evalCfg *EvalConfig
}

func newParseConfig(opts []Option) parseConfig {
	cfg := parseConfig{
		evalCfg: conf.CreateNew(),
		docInfo: DocumentInfo{
			FileName:      "",
			ByteOffset:    0,
			StartPosition: parsetypes.NewEmptyPosition(),
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	for name := range cfg.evalCfg.Disabled {
		delete(cfg.evalCfg.Builtins, name)
	}

	cfg.evalCfg.Check()
	return cfg
}
