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
)

type docBlocks struct {
	vars    *Vars
	tasks   map[string]*hcl.Block
	params  map[string]*Param
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
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("hcl file body is not %T (got %T)", body, f.Body)
	}

	version, err := extractVersion(body)
	if err != nil {
		return nil, err
	}

	importList, err := extractImports(body, p.projCtx)
	if err != nil {
		return nil, err
	}

	_, err = p.traverseBlocks(body.Blocks)
	if err != nil {
		return nil, err
	}

	return &Spec{
		Version: version,
		Imports: importList,
	}, nil
}

func (p *Parser) traverseBlocks(blocks hclsyntax.Blocks) (*docBlocks, hcl.Diagnostics) {
	b := &docBlocks{
		tasks:  map[string]*hcl.Block{},
		params: map[string]*Param{},
	}
	for _, block := range blocks {
		switch block.Type {
		case varsBlockName:
			if b.vars != nil {
				return nil, newDiagnosticError(block.DefRange(),
					"duplicate vars block. Previous declaration was on line %d",
					b.vars.Range.Start.Line)
			}

			vars, err := ParseVarsBlock(block.AsHCLBlock())
			if err != nil {
				return nil, err
			}

			b.vars = vars
		case paramBlockName:
			param, err := ParamFromBlock(block, p.projCtx)
			if err != nil {
				return nil, err
			}
			b.params[param.Name] = param
		default:
			b.unknown = append(b.unknown, block)
		}
	}

	return b, nil
}

func extractImports(body *hclsyntax.Body, ctx *hcl.EvalContext) ([]string, error) {
	importList, err := extractListAttr[string]("imports", body, ctx)
	if err != nil {
		return nil, err
	}

	return importList, nil
}

func extractVersion(body *hclsyntax.Body) (uint, error) {
	version, ok, err := extractAttr[uint]("version", body, nil)
	if err != nil {
		return 0, err
	}

	if !ok {
		return 0, ErrVersionMissing
	}

	if version > maxHclVersion {
		return 0, ErrUnsupportedVersion
	}

	return version, nil
}
