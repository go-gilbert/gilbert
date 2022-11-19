package spec

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

const maxHclVersion = 1

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

	return &Spec{
		Version: version,
		Imports: importList,
	}, nil
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
	//verAttr, ok := body.Attributes["version"]
	//if !ok {
	//	return 0, ErrVersionMissing
	//}
	//
	//ver, err := verAttr.Expr.Value(nil)
	//if err != nil {
	//	return 0, fmt.Errorf("failed to parse version attribute: %w", err)
	//}
	//
	//var version uint
	//if err := gocty.FromCtyValue(ver, &version); err != nil {
	//	return 0, fmt.Errorf("invalid version attribute: %w", err)
	//}
	//
	return version, nil
}

func extractListAttr[T any](name string, body *hclsyntax.Body, ctx *hcl.EvalContext) ([]T, error) {
	val, err := getAttr(name, body, ctx)
	if err != nil {
		return nil, err
	}

	if val == cty.NilVal {
		return nil, err
	}

	out, err := ctyTupleToSlice[T](val)
	if err != nil {
		return nil, fmt.Errorf("invalid %s attribute: %w", name, err)
	}

	return out, nil
}

func extractAttr[T any](name string, body *hclsyntax.Body, ctx *hcl.EvalContext) (T, bool, error) {
	var out T
	val, err := getAttr(name, body, ctx)
	if err != nil {
		return out, false, err
	}

	if val == cty.NilVal {
		return out, false, nil
	}

	if err := gocty.FromCtyValue(val, &out); err != nil {
		return out, true, fmt.Errorf("invalid %s attribute: %w", name, err)
	}

	return out, true, nil
}

func getAttr(name string, body *hclsyntax.Body, ctx *hcl.EvalContext) (cty.Value, error) {
	attr, ok := body.Attributes[name]
	if !ok {
		return cty.NilVal, nil
	}

	val, err := attr.Expr.Value(ctx)
	if err != nil {
		return cty.NilVal, fmt.Errorf("failed to parse %s attribute: %w", name, err)
	}

	return val, nil
}

func ctyTupleToSlice[T any](val cty.Value) ([]T, error) {
	var out []T
	if !val.Type().IsTupleType() {
		err := gocty.FromCtyValue(val, out)
		return out, err
	}

	out = make([]T, val.Type().Length())
	for i := range out {
		elem := val.Index(cty.NumberIntVal(int64(i)))
		if err := gocty.FromCtyValue(elem, &out[i]); err != nil {
			return nil, fmt.Errorf("invalid element at index %d, %w", i, err)
		}
	}

	return out, nil
}
