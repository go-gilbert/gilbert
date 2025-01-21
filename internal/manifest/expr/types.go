package expr

// CommandProcessor handles processing of shell command expressions in manifest file.
type CommandProcessor interface {
	// EvalCommand runs a shell command and returns output result as bytes.
	EvalCommand(command string) ([]byte, error)
}

// ValueResolver resolves variables mentioned in expressions.
type ValueResolver interface {
	GetValue(varName string) (string, bool)
}

type EvalContext struct {
	CommandProcessor CommandProcessor
	Values           ValueResolver
}
