package yamlloader

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/goccy/go-yaml/ast"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
)

var nestedJobSchema *ObjectVisitor[manifest.Job]

func init() {
	// hack for recursive job schema reference.
	nestedJobSchema = jobSchema
}

var jobGroupsSchema = Map(Pointer(jobGroupSchema)).
	CollectDoc(func(_ context.Context, fi FieldInfo, g *manifest.JobGroup) *manifest.JobGroup {
		g.Name = fi.Key
		g.Doc = fi.Doc
		return g
	})

var jobGroupSchema = Struct(
	Field(
		"inputs",
		inputsSchema,
		func(_ context.Context, dst *manifest.JobGroup, v manifest.Inputs) error {
			dst.Inputs = v
			return nil
		},
	),
	Field(
		"steps",
		List(jobSchema),
		func(_ context.Context, dst *manifest.JobGroup, v []manifest.Job) error {
			dst.Jobs = v
			return nil
		},
	).Required(),
)

var jobSchema = Struct(
	Field(
		"action", String(),
		func(ctx context.Context, dst *manifest.Job, val string) error {
			return setJobTarget(ctx, dst, manifest.JobKindAction, val)
		},
	).CheckNode(setJobTargetValueLoc),
	Field(
		"mixin", String(),
		func(ctx context.Context, dst *manifest.Job, val string) error {
			return setJobTarget(ctx, dst, manifest.JobKindMixin, val)
		},
	).CheckNode(setJobTargetValueLoc),
	Field(
		"async", Bool(),
		func(_ context.Context, dst *manifest.Job, v bool) error {
			dst.Async = v
			return nil
		},
	),
	Field(
		"delay", Duration(),
		func(_ context.Context, dst *manifest.Job, v time.Duration) error {
			dst.Delay = v
			return nil
		},
	),
	Field(
		"timeout", Duration(),
		func(_ context.Context, dst *manifest.Job, v time.Duration) error {
			dst.Timeout = v
			return nil
		},
	),
	Field(
		"if",
		Transform(
			String(), func(ctx context.Context, n ast.Node, val string) (*manifest.LazyValue, error) {
				ldCtx, err := getLoaderContext(ctx)
				if err != nil {
					return nil, err
				}

				ex, diag := expressionFromAnyNode(ldCtx.filePath, n, val)
				if diag != nil {
					return nil, diag
				}

				if !ex.Evaluable() {
					return nil, errors.New("value should be an evaluable expression")
				}

				loc, err := buildRefLocation(ctx, n)
				if err != nil {
					return nil, err
				}

				return &manifest.LazyValue{
					Location: loc,
					Value: manifest.AnySpec{
						BindingSpec: &manifest.BindingSpec{
							Location: *loc,
							Expr:     ex,
						},
					},
				}, nil
			},
		),
		func(_ context.Context, dst *manifest.Job, v *manifest.LazyValue) error {
			dst.Condition = v
			return nil
		},
	),
	Field(
		"strategy", strategySchema,
		func(_ context.Context, dst *manifest.Job, v manifest.ExecStrategy) error {
			dst.Strategy = v
			return nil
		},
	),
	Field(
		"with",
		Map(lazyValueVisitor{}),
		func(_ context.Context, dst *manifest.Job, v map[string]*manifest.LazyValue) error {
			// fmt.Println("set with: ", dst.Handler, v)
			dst.Args = v
			return nil
		},
	),
	Field(
		"on",
		Map(
			List(
				Selector(
					func(_ context.Context, _ ast.Node) (ValueVisitor[manifest.Job], error) {
						return nestedJobSchema, nil
					},
				),
			),
		),
		func(_ context.Context, dst *manifest.Job, v map[string][]manifest.Job) error {
			dst.Hooks = v
			return nil
		},
	),
).
	Validation(func(ctx context.Context, n ast.Node, dst *manifest.Job) error {
		if dst.Kind == manifest.JobKindUnknown {
			return errors.New(`missing action target, please set either "action" or "mixin" field`)
		}

		loc, err := buildRefLocation(ctx, n)
		if err != nil {
			return err
		}

		dst.Location = *loc
		return nil
	})

func setJobTargetValueLoc(ctx context.Context, n *ast.MappingValueNode, j *manifest.Job) {
	loc, err := buildRefLocation(ctx, n.Value)
	if err != nil {
		return
	}

	j.Handler.Location = loc
}

func setJobTarget(_ context.Context, j *manifest.Job, targetType manifest.JobKind, name string) error {
	if j.Kind != manifest.JobKindUnknown {
		return errors.New(`only one of "action" or "mixin" fields can be set`)
	}

	j.Kind = targetType
	switch targetType {
	case manifest.JobKindAction:
		h, err := manifest.SplitActionName(name)
		if err != nil {
			return err
		}

		h.Location = j.Handler.Location
		j.Handler = h
		return nil
	case manifest.JobKindMixin, manifest.JobKindTask:
		if err := manifest.ValidateTaskName(name); err != nil {
			return err
		}

		j.Handler.Name = name
		return nil
	}

	return fmt.Errorf("unknown job kind: %v", targetType)
}

var (
	matrixSchema = OrderedMap(lazyArrayVisitor{}, func(k string, v *manifest.LazyValue) manifest.MatrixParam {
		return manifest.MatrixParam{
			Key:    k,
			Values: v,
		}
	})

	matRuleSchema = Map(Transform(AnyScalar(), func(ctx context.Context, n ast.Node, v any) (*manifest.MatrixMatchValue, error) {
		if !parsetypes.IsScalar(v) {
			return nil, parsetypes.NewAnnotatedError(
				errors.New("exclude rule value should be a primitive"),
				"value is not comparable",
			)
		}

		loc, err := buildRefLocation(ctx, n)
		if err != nil {
			return nil, err
		}

		return &manifest.MatrixMatchValue{
			Location: loc,
			Value:    v,
		}, nil
	}))

	strategySchema = Struct(
		Field("max-parallel", UInt[uint](), func(_ context.Context, dst *manifest.ExecStrategy, v uint) error {
			if v == 0 {
				// TODO: throw warning instead of error
				return errors.New("value should be greater than zero")
			}

			dst.MaxParallel = int(v)
			return nil
		}),
		Field("continue-on-error", Bool(), func(_ context.Context, dst *manifest.ExecStrategy, v bool) error {
			dst.ContinueOnError = v
			return nil
		}),
		Field(
			"matrix", matrixSchema,
			func(_ context.Context, dst *manifest.ExecStrategy, v []manifest.MatrixParam) error {
				if len(v) == 0 {
					return errors.New("empty matrix")
				}

				dst.SetMatrixParams(v)
				return nil
			},
		).Required(),
		Field(
			"exclude",
			List(matRuleSchema),
			func(_ context.Context, dst *manifest.ExecStrategy, v []manifest.MatrixMatchRule) error {
				dst.Exclude = v
				return nil
			}).Validation(func(ctx context.Context, es *manifest.ExecStrategy) error {
			if len(es.Exclude) == 0 {
				return parsetypes.Warningf(`redundant "exclude" field`)
			}

			if len(es.MatrixKeys) == 0 {
				return errors.New(`missing "matrix" field`)
			}

			// find keys which are in
			badKeys := map[string]struct{}{}
			for _, e := range es.Exclude {
				for k := range e {
					if _, ok := es.MatrixKeys[k]; !ok {
						badKeys[k] = struct{}{}
					}
				}
			}

			if len(badKeys) == 0 {
				return nil
			}

			return parsetypes.Warningf(
				`"exclude" block contains parameters not present in "matrix"`,
			).WithNote("redundant keys: %q", slices.Collect(maps.Keys(badKeys)))
		}),
	).Constructor(func(_ context.Context, es *manifest.ExecStrategy) error {
		es.MaxParallel = 1
		return nil
	})
)

func buildRefLocation(ctx context.Context, n ast.Node) (*manifest.ReferenceLocation, error) {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return nil, err
	}

	return buildRefLocationWithCtx(c, n), nil
}

func buildRefLocationWithCtx(c *loaderContext, n ast.Node) *manifest.ReferenceLocation {
	rng, offset := GetNodeRange(n)
	return &manifest.ReferenceLocation{
		FileName: c.filePath,
		Range:    rng,
		Offset:   offset,
	}
}
