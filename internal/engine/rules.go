package engine

import (
	"context"
	"errors"

	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

// shouldRunJob returns whetner job should be executed by checking "if" and other conditional attributes.
func shouldRunJob(ctx context.Context, ep expr.EvalParams, j *manifest.Job) (bool, *parsetypes.Diagnostic) {
	if j.Condition == nil {
		return true, nil
	}

	loc := j.Condition.Location
	if loc == nil {
		loc = &j.Location
	}

	v, err := j.Condition.Expand(ctx, ep)
	if err != nil {
		if d, ok := parsetypes.DiagnosticFromError(err); ok {
			return false, d
		}

		return false, &parsetypes.Diagnostic{
			FileName: loc.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    loc.Range,
			Offset:   loc.Offset,
			Note:     err.Error(),
			Err:      errors.New(`failed to evaluate expression inside "if" property`),
		}
	}

	b, err := parsetypes.AnyToBool(v)
	if err != nil {
		return false, &parsetypes.Diagnostic{
			FileName: loc.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    loc.Range,
			Offset:   loc.Offset,
			Note:     err.Error(),
			Err:      errors.New(`invalid value in "if" property`),
		}
	}

	return b, nil
}
