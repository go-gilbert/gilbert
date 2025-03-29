package manifest

const DefaultDelimiter = ","

type InputBinding struct {
	// EnvVarName is environment variable to use by default if value is not defined.
	EnvVarName string

	// FlagName is custon command-line flag name to use for a value.
	FlagName string

	// Delimiter is list item delimiter character in flag or env var string.
	Delimiter string
}

func (b *InputBinding) DelimiterOrDefault() string {
	if b == nil || b.Delimiter == "" {
		return DefaultDelimiter
	}

	return b.Delimiter
}

type Inputs = map[string]*InputDefinition

// InputDefinition is workflow or job input parameter definition.
type InputDefinition struct {
	DocHeader
	Location     ReferenceLocation
	Schema       TypeSchema
	Binding      *InputBinding
	DefaultValue *TypedLazyValue
}
