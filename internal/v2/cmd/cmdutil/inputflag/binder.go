package inputflag

import (
	"context"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/spf13/cobra"
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
	EvalContext expr.EvalContext
	EnvVars     map[string]string
	Scope       *scope.Scope
	Diagnostics *DiagnosticsCollector
}

// InputFlagsBinder mounts workflow inputs as cobra command flags.
//
// Flag values are parsed and validated against input definition schema.
// Parsed input values are mounted into a scope passed in InputBindingOpts.
type InputFlagsBinder struct {
	ctx      context.Context
	logger   *log.Logger
	opts     InputBindingOpts
	diags    *DiagnosticsCollector
	bindings []inputFlagBinding
}

func NewInputFlagsBinder(ctx context.Context, logger *log.Logger, opts InputBindingOpts) *InputFlagsBinder {
	return &InputFlagsBinder{
		logger: logger,
		opts:   opts,
		ctx:    ctx,
		diags:  opts.Diagnostics,
	}
}

func (b *InputFlagsBinder) bindFlag(input *manifest.InputDefinition, cmd *cobra.Command, isGlobal bool) error {
	inputCtx := inputFlagContext{
		evalContext:      b.opts.EvalContext,
		envVars:          b.opts.EnvVars,
		dstScope:         b.opts.Scope,
		inputDiagnostics: b.diags,
	}

	binding, err := flagBindingFromInput(b.logger, input, inputCtx)
	if err != nil {
		return err
	}

	binding.initDefaultValue(b.ctx)
	b.bindings = append(b.bindings, binding)

	flagName := binding.flagName()
	flagDoc := binding.getDoc(isGlobal)
	if isGlobal {
		cmd.PersistentFlags().Var(binding, flagName, flagDoc)
	} else {
		cmd.Flags().Var(binding, flagName, flagDoc)
	}

	if binding.isRequired() {
		if isGlobal {
			err = cmd.MarkPersistentFlagRequired(flagName)
		} else {
			err = cmd.MarkFlagRequired(flagName)
		}
	}

	if err != nil {
		err = fmt.Errorf("failed to bind input to flag %q: %w", flagName, err)
	}

	return err
}

// BindGlobalInput binds given global input parameter as a cobra persistent command flag.
func (b *InputFlagsBinder) BindGlobalInput(input *manifest.InputDefinition, cmd *cobra.Command) error {
	return b.bindFlag(input, cmd, true)
}

// BindTaskInput binds given task input parameter as a cobra command flag.
func (b *InputFlagsBinder) BindTaskInput(input *manifest.InputDefinition, cmd *cobra.Command) error {
	return b.bindFlag(input, cmd, false)
}

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
