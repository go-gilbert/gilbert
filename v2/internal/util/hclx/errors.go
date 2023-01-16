package hclx

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type DiagnosticOption = func(d *hcl.Diagnostic)

// WithDetail adds detailed message to diagnostic
func WithDetail(msg string, args ...any) DiagnosticOption {
	return func(d *hcl.Diagnostic) {
		d.Detail = formatMsgWithArgs(msg, args)
	}
}

// WithSummary fills diagnostic short summary.
func WithSummary(msg string, args ...any) DiagnosticOption {
	return func(d *hcl.Diagnostic) {
		d.Summary = formatMsgWithArgs(msg, args)
	}
}

// WithContext fills diagnostic context range.
func WithContext(rng hcl.Range) DiagnosticOption {
	return func(d *hcl.Diagnostic) {
		d.Context = &rng
	}
}

// WithSubject fills error subject range
func WithSubject(rng hcl.Range) DiagnosticOption {
	return func(d *hcl.Diagnostic) {
		d.Subject = &rng
	}
}

// NewDiagnostic returns hcl.Diagnostics with a single error.
func NewDiagnostic(rng hcl.Range, opt DiagnosticOption, opts ...DiagnosticOption) *hcl.Diagnostic {
	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Subject:  &rng,
	}

	opt(diag)
	if len(opts) == 0 {
		return diag
	}

	for _, opt := range opts {
		opt(diag)
	}

	return diag
}

// NewDiagnosticError creates a new diagnostic error.
//
// Deprecated. Use
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

func formatMsgWithArgs(msg string, args []any) string {
	if len(args) == 0 {
		return msg
	}

	return fmt.Sprintf(msg, args...)
}
