package inputflag

import (
	"context"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/spf13/pflag"
)

// inputFlagBinding is interface to implement binding from job file input into command-line flags.
//
// Command-line flag value is decoded using input definition and written into a passed scope.
type inputFlagBinding interface {
	pflag.Value

	// getDoc returns flag documentation.
	getDoc(isGlobal bool) string

	// flagName returns command-line flag to attach a binding.
	flagName() string

	// isRequired returns whether input value is mandatory.
	isRequired() bool

	// initDefaultValue initializes flag default value (if declared in input definition).
	initDefaultValue(ctx context.Context)

	// error returns an error if there was an issue while decoding default flag value.
	error() error
}

type InputBindingOpts struct {
	EvalParams  expr.EvalParams
	EnvVars     map[string]string
	Scope       *scope.Scope
	Diagnostics *DiagnosticsCollector
}

// InputFlagsBinder mounts workflow inputs as cobra command flags.
//
// Flag values are parsed and validated against input definition schema.
// Parsed input values are mounted into a scope passed in InputBindingOpts.
type InputFlagsBinder struct {
	logger *log.Logger
	opts   InputBindingOpts
	diags  *DiagnosticsCollector
}

func NewInputFlagsBinder(logger *log.Logger, opts InputBindingOpts) *InputFlagsBinder {
	return &InputFlagsBinder{
		logger: logger,
		opts:   opts,
		diags:  opts.Diagnostics,
	}
}

// BindInputsToFlagSet constructs flag set from inputs list.
//
// Returns list of required flags.
func (b *InputFlagsBinder) BindInputsToFlagSet(ctx context.Context, fset *pflag.FlagSet, isGlobal bool, inputs manifest.Inputs) ([]string, error) {
	inputCtx := inputFlagContext{
		evalParams:       b.opts.EvalParams,
		envVars:          b.opts.EnvVars,
		dstScope:         b.opts.Scope,
		inputDiagnostics: b.diags,
	}

	requiredFlags := make([]string, 0, len(inputs)/2)
	for _, input := range inputs {
		binding, err := flagBindingFromInput(b.logger, input, inputCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to bind input %q to flag: %w", input.Name, err)
		}

		binding.initDefaultValue(ctx)
		flagName := binding.flagName()
		flagDoc := binding.getDoc(isGlobal)
		if binding.isRequired() {
			requiredFlags = append(requiredFlags, flagName)
		}

		fset.Var(binding, flagName, flagDoc)
	}

	return requiredFlags, nil
}

// BindTaskInput binds given task input parameter as a cobra command flag.
func flagBindingFromInput(logger *log.Logger, inputDef *manifest.InputDefinition, inputCtx inputFlagContext) (inputFlagBinding, error) {
	if inputDef.Schema.Type.IsList() {
		return newListInputFlagBinding(logger, inputDef, inputCtx), nil
	}

	if inputDef.Schema.Type.IsComplex() {
		// objects aren't supported (yet)
		return nil, fmt.Errorf("cannot bind input %q to a flag: complex types are not supported", inputDef.Name)
	}

	return newScalarInputFlagBinding(logger, inputDef, inputCtx), nil
}
