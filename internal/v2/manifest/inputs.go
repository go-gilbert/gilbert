package manifest

const DefaultDelimiter = ","

type InputBindingLocations struct {
	EnvVarName *ReferenceLocation
	FlagName   *ReferenceLocation
}

type InputBinding struct {
	// Location holds reference location for input binding values.
	//
	// Used for error reporting.
	Location InputBindingLocations

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
	Optional     bool
}

// IsRequired returns whether input parameter is required.
//
// Input is considered as required when there is no default value, not marked as optional and it's not boolean.
func (def *InputDefinition) IsRequired() bool {
	if def.DefaultValue != nil {
		return false
	}

	isOptional := def.Schema.Type == ValueTypeBool || def.Optional
	return !isOptional
}
