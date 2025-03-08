package expr

// CommandProcessor handles processing of shell command expressions in manifest file.
type CommandProcessor interface {
	// EvalCommand runs a shell command and returns output result as bytes.
	EvalCommand(command string) ([]byte, error)
}

// ValueResolver resolves variables mentioned in expressions.
type ValueResolver interface {
	// ValueByName returns a value by variable name.
	ValueByName(varName string) (string, bool)

	// Values returns a raw value of a container holding all values.
	Values() any
}

type EvalContext struct {
	CommandProcessor CommandProcessor
	Env              ValueResolver
}
