package golang

import (
	"context"
	"errors"
	"strings"

	"github.com/go-gilbert/gilbert/internal/v2/engine"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	. "github.com/go-gilbert/gilbert/internal/v2/manifest/argschema"
)

type buildActionArgs struct {
	CGoEnabled bool
	BuildVCS   bool

	PackageName string
	Output      string
	WorkDir     string
	OS          string
	Arch        string
	BuildMode   string
	BuildTags   []string
	AsmFlags    []string
	LinkerFlags []string
	GCCGOFlags  []string
	GCFlags     []string

	PackageVars map[string]string
	Env         map[string]string
}

// TODO: validation
var buildSchema = Struct(
	StringField("package", func(dst *buildActionArgs, val string) error {
		val = strings.TrimSpace(val)
		if val == "" {
			return errors.New("package name is required")
		}

		dst.PackageName = val
		return nil
	}).Required(),
	StringField("output", func(dst *buildActionArgs, val string) error {
		dst.Output = val
		return nil
	}),
	StringField("workdir", func(dst *buildActionArgs, val string) error {
		dst.WorkDir = val
		return nil
	}),
	StringField("os", func(dst *buildActionArgs, val string) error {
		dst.OS = val
		return nil
	}),
	StringField("arch", func(dst *buildActionArgs, val string) error {
		dst.Arch = val
		return nil
	}),
	StringField("buildmode", func(dst *buildActionArgs, val string) error {
		dst.BuildMode = val
		return nil
	}),
	BoolField("buildvcs", func(dst *buildActionArgs, val bool) error {
		dst.BuildVCS = val
		return nil
	}),
	BoolField("cgo", func(dst *buildActionArgs, val bool) error {
		dst.CGoEnabled = val
		return nil
	}),
	ListField("tags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.BuildTags = val
		return nil
	}),
	ListField("gccgoflags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.GCCGOFlags = val
		return nil
	}),
	ListField("gcflags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.GCFlags = val
		return nil
	}),
	ListField("ldflags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.LinkerFlags = val
		return nil
	}),
	ListField("asmflags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.AsmFlags = val
		return nil
	}),
	Field("package_vars", Dict(AnyString), func(_ VisitParams, dst *buildActionArgs, val map[string]string) error {
		dst.PackageVars = val
		return nil
	}),
	Field("env", Dict(AnyString), func(_ VisitParams, dst *buildActionArgs, val map[string]string) error {
		dst.Env = val
		return nil
	}),
)

var _ engine.ActionHandler = (*BuildActionHandler)(nil)

type BuildActionHandler struct {
	logger *log.Logger
	shell  *engine.Shell

	args buildActionArgs
}

func NewBuildActionHandler(ctx context.Context, ref manifest.JobHandlerRef, params engine.ActionParams) engine.HandlerResult {
	args, diags := MapArgsToStruct(ctx, params, buildSchema)
	if diags.HasError() {
		return engine.NewBadActionParamsResult(ref, diags)
	}

	h := &BuildActionHandler{
		logger: params.Logger,
		shell:  params.Shell,
		args:   args,
	}
	return engine.NewHandlerResult(h, diags)
}

func (b *BuildActionHandler) HandleAction(ctx context.Context) error {
	return errors.New("not implemented")
}
