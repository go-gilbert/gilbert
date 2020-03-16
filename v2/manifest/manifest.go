package manifest

import (
	"io/ioutil"
	"path/filepath"

	"github.com/go-gilbert/gilbert/v2/manifest/context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

const (
	propImports = "imports"

	DefaultFileName = "gilbert.hcl"
)

var startPos = hcl.Pos{Line: 1, Column: 1}

type Manifest struct {
	src      []byte
	ctx      *hcl.EvalContext
	FileName string
	Location string
	Plugins  []string
	Imports  []string
	Tasks    Tasks
	Mixins   Mixins
}

func (m *Manifest) FilePath() string {
	return filepath.Join(m.Location, m.FileName)
}

func FromFile(fileName string, parentCtx *hcl.EvalContext) (*Manifest, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	f, diags := hclsyntax.ParseConfig(data, fileName, startPos)
	if diags != nil {
		return nil, NewError(fileName, "", data, diags)
	}

	// prepare root parse context
	body := f.Body.(*hclsyntax.Body)
	ctx := parentCtx.NewChild()
	ctx.Variables = make(map[string]cty.Value, len(body.Attributes))
	ctx.Functions = context.GetDefaultFunctions()

	// extract all global variables
	if diags := appendAttrsToContext(body.Attributes, ctx); diags != nil {
		return nil, NewError(fileName, "", f.Bytes, diags)
	}

	// extract tasks and mixins
	tasks, mixins, diags := extractTasksAndMixins(ctx, body.Blocks)
	if diags != nil {
		return nil, NewError(fileName, "", f.Bytes, diags)
	}

	return &Manifest{
		ctx:      ctx,
		src:      data,
		FileName: filepath.Base(fileName),
		Location: filepath.Dir(fileName),
		Tasks:    tasks,
		Mixins:   mixins,
	}, nil
}

func (m *Manifest) IncludeParent(parent *Manifest) {
	// TODO: add implementation
	panic("not implemented")
}
