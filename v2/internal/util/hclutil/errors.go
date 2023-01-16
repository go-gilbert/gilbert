package hclutil

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// NewDiagnosticError creates a new diagnostic error.
func NewDiagnosticError(rng hcl.Range, msg string, args ...any) hcl.Diagnostics {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     msg,
			Detail:      msg,
			Subject:     &rng,
			Context:     &rng,
			Expression:  nil,
			EvalContext: nil,
			Extra:       nil,
		},
	}
}

// WrapDiagnostic wraps diagnostics summary with custom message prefix.
func WrapDiagnostic(parent hcl.Diagnostics, msg string, args ...any) hcl.Diagnostics {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	newDiags := make(hcl.Diagnostics, len(parent))
	for i, diag := range parent {
		if diag.Summary != "" {
			diag.Summary = fmt.Sprint(msg, ": ", diag.Summary)
		} else {
			diag.Summary = msg
		}

		newDiags[i] = diag
	}

	return newDiags
}
