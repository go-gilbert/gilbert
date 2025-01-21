package expr

import "fmt"

// Parser is generic interface to implement template expression parsing for different language spec versions.
type Parser interface {
	// ReadString evaluates expressions inside the string and returns processed result.
	ReadString(ctx EvalContext, str string) (string, error)

	// ContainsExpression checks whether passed string contains template expressions.
	ContainsExpression(str string) bool

	// ReadExpression evaluates a single expression.
	ReadExpression(ctx EvalContext, expr []byte) ([]byte, error)
}

// GetParser returns expression parser for a specific language spec version.
func GetParser(version string) (Parser, error) {
	if version == "2" {
		return NewSpecV2Parser(), nil
	}

	return nil, fmt.Errorf("unsupported language version: %s", version)
}
