package cmdutil

import (
	"context"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/spf13/cobra"
)

type InputBindingOpts struct {
	EvalContext expr.EvalContext
	EnvVars     map[string]string
	Scope       *scope.Scope
}

// InputFlagsBinder mounts workflow inputs as cobra command flags.
//
// Flag values are parsed and validated against input definition schema.
// Parsed input values are mounted into a scope passed in InputBindingOpts.
type InputFlagsBinder struct {
	ctx      context.Context
	opts     InputBindingOpts
	diags    inputDiagnostics
	bindings []*inputFlagBinding
}

func NewInputFlagsBinder(ctx context.Context, opts InputBindingOpts) *InputFlagsBinder {
	return &InputFlagsBinder{
		opts: opts,
		ctx:  ctx,
	}
}

func (b *InputFlagsBinder) bindFlag(input *manifest.InputDefinition, cmd *cobra.Command, isGlobal bool) error {
	inputCtx := inputFlagContext{
		evalContext:      b.opts.EvalContext,
		envVars:          b.opts.EnvVars,
		dstScope:         b.opts.Scope,
		inputDiagnostics: &b.diags,
	}

	binding := newInputFlagBinding(input, inputCtx)
	binding.initDefaultValue(b.ctx)
	b.bindings = append(b.bindings, binding)

	flagName := binding.flagName()
	flagDoc := binding.getDoc(isGlobal)
	if isGlobal {
		cmd.PersistentFlags().Var(binding, flagName, flagDoc)
	} else {
		cmd.Flags().Var(binding, flagName, flagDoc)
	}

	var err error
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
