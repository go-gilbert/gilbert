package yamlloader

import (
	"context"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
)

var jobSchema = Struct[manifest2.Job]()

var jobGroupSchema = Struct[manifest2.JobGroup](
	Field[manifest2.JobGroup, manifest2.Inputs](
		"inputs",
		inputsSchema,
		func(_ context.Context, dst *manifest2.JobGroup, v manifest2.Inputs) error {
			dst.Inputs = v
			return nil
		},
	),
	Field[manifest2.JobGroup, []manifest2.Job](
		"steps",
		List[manifest2.Job](jobSchema),
		func(_ context.Context, dst *manifest2.JobGroup, v []manifest2.Job) error {
			dst.Jobs = v
			return nil
		},
	).Required(),
)
