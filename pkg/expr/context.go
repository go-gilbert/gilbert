package expr

import (
	"context"
	"errors"
)

// CommandProcessor handles processing of shell command expressions in manifest file.
type CommandProcessor interface {
	// EvalCommand runs a shell command and returns output result as bytes.
	EvalCommand(ctx context.Context, command string) ([]byte, error)
}

// ValueResolver resolves variables mentioned in expressions.
type ValueResolver interface {
	// Values returns a raw value of a container holding all values.
	Values() (map[string]any, error)
}

type EvalParams struct {
	CommandProcessor CommandProcessor
	Env              ValueResolver
}

// NoopEvalParams are params with empty env and command processor which always returns an error.
//
// Used as a placeholder to resolve a value where dynamic expressions are disallowed.
var NoopEvalParams = EvalParams{
	CommandProcessor: NoopCommandProcessor{},
	Env:              NoopValueResolver{},
}

type NoopCommandProcessor struct{}

func (NoopCommandProcessor) EvalCommand(context.Context, string) ([]byte, error) {
	return nil, errors.New("shell invocation not allowed in this context")
}

type NoopValueResolver struct{}

func (NoopValueResolver) Values() (map[string]any, error) {
	return nil, errors.New("expressions are not allowed in this context")
}
