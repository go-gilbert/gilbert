package spec

import (
	"strings"

	"github.com/go-gilbert/gilbert/v2/internal/util/options"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const (
	nameLabelIndex        = 0
	descriptionLabelIndex = 1

	headerKeyAttr = "key"
)

type headerParseOpts struct {
	optionalName bool
	withKey      bool
}

type ParseHeaderOption = options.Option[headerParseOpts]

// WithOptionalBlockName makes block name optional.
var WithOptionalBlockName = ParseHeaderOption(func(opts *headerParseOpts) {
	opts.optionalName = true
})

// WithBlockKey option enables obtain of "key" attribute.
var WithBlockKey = func(readBlockKey bool) ParseHeaderOption {
	return func(opts *headerParseOpts) {
		opts.optionalName = readBlockKey
	}
}

type BlockHeader struct {
	Key         string
	Name        string
	Description string
}

func ParseHeader(block *hclsyntax.Block, ctx *hcl.EvalContext, opts ...ParseHeaderOption) (BlockHeader, hcl.Diagnostics) {
	var opt headerParseOpts
	options.Apply(&opt, opts)

	name := lookupLabel(block.Labels, nameLabelIndex)
	if name == "" && !opt.optionalName {
		return BlockHeader{}, newDiagnosticError(block.DefRange(),
			"missing name label in %s block", block.Type)
	}

	header := BlockHeader{
		Name:        name,
		Description: lookupLabel(block.Labels, descriptionLabelIndex),
	}

	keyAttr, hasKey := block.Body.Attributes[headerKeyAttr]
	if !opt.withKey || !hasKey {
		return header, nil
	}

	key, err := unmarshalAttr[string](keyAttr.AsHCLAttribute(), ctx)
	header.Key = key
	return header, err
}

func lookupLabel(labels []string, index int) string {
	if len(labels) > index {
		return strings.TrimSpace(labels[index])
	}

	return ""
}
