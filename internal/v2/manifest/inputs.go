package manifest

type InputBinding struct {
	// EnvVarName is environment variable to use by default if value is not defined.
	EnvVarName string

	// FlagName is custon command-line flag name to use for a value.
	FlagName string
}

type Inputs = map[string]*InputDefinition

type InputDefinition struct {
	DocHeader
	Location     ReferenceLocation
	Type         TypeSchema
	Binding      *InputBinding
	DefaultValue *TypedLazyValue
}
