package expr

import "context"

// CommandProcessor handles processing of shell command expressions in manifest file.
type CommandProcessor interface {
	// EvalCommand runs a shell command and returns output result as bytes.
	EvalCommand(ctx context.Context, command string) ([]byte, error)
}

// ValueResolver resolves variables mentioned in expressions.
type ValueResolver interface {
	// ValueByName returns a value by variable name.
	//
	// DEPRECATED: getting individial values won't be supported, use Values instead.
	ValueByName(varName string) (string, bool)

	// Values returns a raw value of a container holding all values.
	Values() map[string]any
}

type EvalContext struct {
	CommandProcessor CommandProcessor
	Env              ValueResolver
}
