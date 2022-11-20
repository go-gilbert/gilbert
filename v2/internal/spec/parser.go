package spec

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const maxHclVersion = 1

const (
	varsBlockName  = "vars"
	paramBlockName = "param"
	taskBlockName  = "task"
	mixinBlockName = "mixin"

	versionAttr = "version"
	importsAttr = "imports"
)

type docBlocks struct {
	vars    *Vars
	params  Params
	tasks   Tasks
	unknown hclsyntax.Blocks
}

type Parser struct {
	project ProjectSpec
	projCtx *hcl.EvalContext
	rootCtx *hcl.EvalContext
}

func NewParser(rootCtx *RootContext, proj ProjectSpec) *Parser {
	rootEvalCtx := &hcl.EvalContext{
		Functions: builtinFunctions,
	}
	rootCtx.Map(rootEvalCtx)

	fileCtx := FileContext{
		Project: proj,
	}

	projCtx := rootEvalCtx.NewChild()
	fileCtx.Map(projCtx)

	return &Parser{
		project: proj,
		rootCtx: rootEvalCtx,
		projCtx: projCtx,
	}
}

func (p *Parser) Parse(data []byte) (*Spec, error) {
	f, err := hclsyntax.ParseConfig(data, p.project.FileName, hcl.InitialPos)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return p.manifestFromHcl(f)
}

func (p *Parser) manifestFromHcl(f *hcl.File) (*Spec, error) {
	doc, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("hcl file body is not %T (got %T)", doc, f.Body)
	}

	header, err := parseHeader(doc, p.projCtx)
	if err != nil {
		return nil, err
	}

	body, err := p.traverseBlocks(doc.Blocks)
	if err != nil {
		return nil, err
	}

	return &Spec{
		Header: header,
		Body:   body,
	}, nil
}

func (p *Parser) traverseBlocks(blocks hclsyntax.Blocks) (Body, hcl.Diagnostics) {
	b := Body{
		Tasks:  make(Tasks, len(blocks)),
		Params: Params{},
	}
	for _, block := range blocks {
		switch block.Type {
		case varsBlockName:
			if b.Vars != nil {
				return b, newDiagnosticError(block.DefRange(),
					"duplicate vars block. Previous declaration was on line %d",
					b.Vars.Range.Start.Line)
			}

			vars, err := ParseVars(block.AsHCLBlock())
			if err != nil {
				return b, err
			}

			b.Vars = vars
		case paramBlockName:
			param, err := ParseParam(block, p.projCtx)
			if err != nil {
				return b, err
			}
			b.Params[param.Name] = param
		case taskBlockName:
			task, err := ParseTask(block, p.projCtx)
			if err != nil {
				return b, err
			}

			b.Tasks[task.Name] = task
		case mixinBlockName:
			panic("unimplemented!")
		default:
			b.Unknown = append(b.Unknown, block)
		}
	}

	return b, nil
}

func parseHeader(body *hclsyntax.Body, ctx *hcl.EvalContext) (head Header, err hcl.Diagnostics) {
	head.Version, err = extractVersion(body.Attributes)
	if err != nil {
		return head, err
	}

	head.Imports, err = extractImports(body.Attributes, ctx)
	if err != nil {
		return head, err
	}

	return head, nil
}

func extractImports(attrs hclsyntax.Attributes, ctx *hcl.EvalContext) ([]string, hcl.Diagnostics) {
	importList, err := extractListAttr[string](importsAttr, attrs, ctx)
	if err != nil {
		return nil, err
	}

	return importList, nil
}

func extractVersion(attrs hclsyntax.Attributes) (uint, hcl.Diagnostics) {
	attr, ok := attrs[versionAttr]
	if !ok {
		return 0, newDiagnosticError(hcl.Range{}, "missing file version")
	}

	version, err := unmarshalAttr[uint](attr.AsHCLAttribute(), nil)
	if err != nil {
		return 0, err
	}

	if version > maxHclVersion {
		return 0, newDiagnosticError(attr.Range(), "unsupported file version (max: %d)", maxHclVersion)
	}

	return version, nil
}
