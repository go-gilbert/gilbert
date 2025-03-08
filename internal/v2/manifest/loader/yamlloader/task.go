package yamlloader

import (
	"context"
	"errors"
	"fmt"
	"time"

	manifest3 "github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

var nestedJobSchema *ObjectVisitor[manifest3.Job]

func init() {
	// hack for recursive job schema reference.
	nestedJobSchema = jobSchema
}

var jobGroupsSchema = Map(Pointer(jobGroupSchema)).
	CollectDoc(func(_ context.Context, fi FieldInfo, g *manifest3.JobGroup) *manifest3.JobGroup {
		g.Name = fi.Key
		g.Doc = fi.Doc
		return g
	})

var jobGroupSchema = Struct(
	Field(
		"inputs",
		inputsSchema,
		func(_ context.Context, dst *manifest3.JobGroup, v manifest3.Inputs) error {
			dst.Inputs = v
			return nil
		},
	),
	Field(
		"steps",
		List(jobSchema),
		func(_ context.Context, dst *manifest3.JobGroup, v []manifest3.Job) error {
			dst.Jobs = v
			return nil
		},
	).Required(),
)

var jobSchema = Struct(
	Field(
		"action", String(),
		func(ctx context.Context, dst *manifest3.Job, val string) error {
			return setJobTarget(ctx, dst, manifest3.JobKindAction, val)
		},
	),
	Field(
		"mixin", String(),
		func(ctx context.Context, dst *manifest3.Job, val string) error {
			return setJobTarget(ctx, dst, manifest3.JobKindMixin, val)
		},
	),
	Field(
		"async", Bool(),
		func(_ context.Context, dst *manifest3.Job, v bool) error {
			dst.Async = v
			return nil
		},
	),
	Field(
		"delay", Duration(),
		func(_ context.Context, dst *manifest3.Job, v time.Duration) error {
			dst.Delay = v
			return nil
		},
	),
	Field(
		"timeout", Duration(),
		func(_ context.Context, dst *manifest3.Job, v time.Duration) error {
			dst.Timeout = v
			return nil
		},
	),
	Field(
		"if",
		Transform(
			String(), func(ctx context.Context, n ast.Node, val string) (*manifest3.LazyValue, error) {
				ex, err := expr.Parse(val)
				if err != nil {
					return nil, err
				}

				if !ex.Evaluable() {
					return nil, errors.New("expected evaluable expression")
				}

				loc, err := buildRefLocation(ctx, n)
				if err != nil {
					return nil, err
				}

				return &manifest3.LazyValue{
					Location: loc,
					Value: manifest3.AnySpec{
						BindingSpec: &manifest3.BindingSpec{
							Location: *loc,
							Expr:     ex,
						},
					},
				}, nil
			},
		),
		func(_ context.Context, dst *manifest3.Job, v *manifest3.LazyValue) error {
			dst.Condition = v
			return nil
		},
	),
	Field(
		"strategy", strategySchema,
		func(_ context.Context, dst *manifest3.Job, v manifest3.ExecStrategy) error {
			dst.Strategy = v
			return nil
		},
	),
	Field(
		"with",
		Map(lazyValueVisitor{}),
		func(_ context.Context, dst *manifest3.Job, v map[string]*manifest3.LazyValue) error {
			dst.Args = v
			return nil
		},
	),
	Field(
		"on",
		Map(
			List(
				Selector(
					func(_ context.Context, _ ast.Node) (ValueVisitor[manifest3.Job], error) {
						return nestedJobSchema, nil
					},
				),
			),
		),
		func(_ context.Context, dst *manifest3.Job, v map[string][]manifest3.Job) error {
			if !dst.Async {
				return errors.New(`"on" block can be used only when "async" is true`)
			}

			dst.Hooks = v
			return nil
		},
	),
).
	Validation(func(ctx context.Context, n ast.Node, dst *manifest3.Job) error {
		if dst.Kind == manifest3.JobKindUnknown {
			return errors.New(`missing action target, please set either "action" or "mixin" field`)
		}

		loc, err := buildRefLocation(ctx, n)
		if err != nil {
			return err
		}

		dst.Location = *loc
		return nil
	})

func setJobTarget(ctx context.Context, j *manifest3.Job, targetType manifest3.JobKind, name string) error {
	if j.Kind != manifest3.JobKindUnknown {
		return errors.New(`only one of "action" or "mixin" fields can be set`)
	}

	j.Kind = targetType
	switch targetType {
	case manifest3.JobKindAction:
		h, err := manifest3.SplitActionName(name)
		if err != nil {
			return err
		}

		j.Handler = h
		return nil
	case manifest3.JobKindMixin, manifest3.JobKindTask:
		if err := manifest3.ValidateTaskName(name); err != nil {
			return err
		}

		j.Handler.Name = name
		return nil
	}

	return fmt.Errorf("unknown job kind: %v", targetType)
}

var strategySchema = Struct(
	Field(
		"matrix", Map(lazyArrayVisitor{}),
		func(_ context.Context, dst *manifest3.ExecStrategy, v map[string]*manifest3.LazyValue) error {
			dst.Matrix = v
			return nil
		},
	),
)

func buildRefLocation(ctx context.Context, n ast.Node) (*manifest3.ReferenceLocation, error) {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return nil, err
	}

	return buildRefLocationWithCtx(c, n), nil
}

func buildRefLocationWithCtx(c *loaderContext, n ast.Node) *manifest3.ReferenceLocation {
	rng, offset := GetNodeRange(n)
	return &manifest3.ReferenceLocation{
		FileName: c.filePath,
		Range:    rng,
		Offset:   offset,
	}
}
