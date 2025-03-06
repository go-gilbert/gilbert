package yamlloader

import (
	"context"
	"errors"
	"time"

	"github.com/go-gilbert/gilbert/internal/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

var nestedJobSchema *ObjectVisitor[manifest2.Job]

func init() {
	// hack for recursive job schema reference.
	nestedJobSchema = jobSchema
}

var jobGroupsSchema = Map(Pointer(jobGroupSchema)).
	CollectDoc(func(_ context.Context, fi FieldInfo, g *manifest2.JobGroup) *manifest2.JobGroup {
		g.Name = fi.Key
		g.Doc = fi.Doc
		return g
	})

var jobGroupSchema = Struct(
	Field(
		"inputs",
		inputsSchema,
		func(_ context.Context, dst *manifest2.JobGroup, v manifest2.Inputs) error {
			dst.Inputs = v
			return nil
		},
	),
	Field(
		"steps",
		List(jobSchema),
		func(_ context.Context, dst *manifest2.JobGroup, v []manifest2.Job) error {
			dst.Jobs = v
			return nil
		},
	).Required(),
)

var jobSchema = Struct(
	Field(
		"action", String(),
		func(_ context.Context, dst *manifest2.Job, val string) error {
			return setJobTarget(dst, manifest2.JobKindAction, val)
		},
	),
	Field(
		"mixin", String(),
		func(_ context.Context, dst *manifest2.Job, val string) error {
			return setJobTarget(dst, manifest2.JobKindMixin, val)
		},
	),
	Field(
		"async", Bool(),
		func(_ context.Context, dst *manifest2.Job, v bool) error {
			dst.Async = v
			return nil
		},
	),
	Field(
		"delay", Duration(),
		func(_ context.Context, dst *manifest2.Job, v time.Duration) error {
			dst.Delay = v
			return nil
		},
	),
	Field(
		"timeout", Duration(),
		func(_ context.Context, dst *manifest2.Job, v time.Duration) error {
			dst.Timeout = v
			return nil
		},
	),
	Field(
		"if",
		Transform(
			String(), func(ctx context.Context, n ast.Node, val string) (*manifest2.LazyValue, error) {
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

				return &manifest2.LazyValue{
					Location: loc,
					Value: manifest2.AnySpec{
						BindingSpec: &manifest2.BindingSpec{
							Location: *loc,
							Expr:     ex,
						},
					},
				}, nil
			},
		),
		func(_ context.Context, dst *manifest2.Job, v *manifest2.LazyValue) error {
			dst.Condition = v
			return nil
		},
	),
	Field(
		"strategy", strategySchema,
		func(_ context.Context, dst *manifest2.Job, v manifest2.ExecStrategy) error {
			dst.Strategy = v
			return nil
		},
	),
	Field(
		"with",
		Map(lazyValueVisitor{}),
		func(_ context.Context, dst *manifest2.Job, v map[string]*manifest2.LazyValue) error {
			dst.Args = v
			return nil
		},
	),
	Field(
		"on",
		Map(
			List(
				Selector(
					func(_ context.Context, _ ast.Node) (ValueVisitor[manifest2.Job], error) {
						return nestedJobSchema, nil
					},
				),
			),
		),
		func(_ context.Context, dst *manifest2.Job, v map[string][]manifest2.Job) error {
			if !dst.Async {
				return errors.New(`"on" block can be used only when "async" is true`)
			}

			dst.Hooks = v
			return nil
		},
	),
).
	Validation(func(ctx context.Context, n ast.Node, dst *manifest2.Job) error {
		if dst.Kind == manifest2.JobKindUnknown {
			return errors.New(`missing action target, please set either "action" or "mixin" field`)
		}

		loc, err := buildRefLocation(ctx, n)
		if err != nil {
			return err
		}

		dst.Location = *loc
		return nil
	})

func setJobTarget(j *manifest2.Job, targetType manifest2.JobKind, name string) error {
	if j.Kind != manifest2.JobKindUnknown {
		return errors.New(`only one of "action" or "mixin" fields can be set`)
	}

	j.Kind = targetType
	j.Name = name
	return nil
}

var strategySchema = Struct(
	Field(
		"matrix", Map(lazyArrayVisitor{}),
		func(_ context.Context, dst *manifest2.ExecStrategy, v map[string]*manifest2.LazyValue) error {
			dst.Matrix = v
			return nil
		},
	),
)

func buildRefLocation(ctx context.Context, n ast.Node) (*manifest2.ReferenceLocation, error) {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return nil, err
	}

	rng, offset := GetNodeRange(n)
	return &manifest2.ReferenceLocation{
		FileName: c.filePath,
		Range:    rng,
		Offset:   offset,
	}, nil
}
