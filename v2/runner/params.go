package runner

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

// CtxParamInjector injects task parameters
// into evaluation context
type CtxParamInjector interface {
	// InjectParameter injects parameter to the context
	InjectParameter(ctx *hcl.EvalContext) error

	// Required returns is if parameter is required
	Required() bool

	// ValueEmpty returns if value is empty
	ValueEmpty() bool
}

// InjectableParams holds set of task parameters from command line
// that should be injected into task's context.
type InjectableParams map[string]CtxParamInjector

// CheckParams checks if all required task parameters are satisfied
//
// Necessary, because cobra.MarkFlagRequired() doesn't seem to work.
func (ip InjectableParams) CheckParams() error {
	for name, ij := range ip {
		// optional params always have non-nil value
		if ij.ValueEmpty() && ij.Required() {
			return fmt.Errorf("missing required task parameter %q", name)
		}
	}

	return nil
}

func (ip InjectableParams) InjectParameters(ctx *hcl.EvalContext) error {
	if len(ip) == 0 {
		return nil
	}
	for name, inj := range ip {
		if err := inj.InjectParameter(ctx); err != nil {
			return fmt.Errorf("failed to inject parameter %q: %w", name, err)
		}
	}
	return nil
}
