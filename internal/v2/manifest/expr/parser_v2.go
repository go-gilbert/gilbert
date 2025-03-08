package expr

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// FIXME: regex fails to capture escapes with "}" chars (might affect winnt envs).

var re = regexp.MustCompile(templateRegEx)

const (
	evalGroup = 1 // match group that contains shell eval expression
	varGroup  = 2 // match group that contains variable print expression

	groupSize = 3 // match group size

	// Regular expression to extract template expressions
	//
	// - Shell-eval string:
	//		$( whoami )
	// - Variable value
	//		${ foo }
	//
	templateRegEx = `(?m)\$\(([^\)]*)\)|\${([^\}]*)}`
)

// expressionMatch is template expression found by expression processor
type expressionMatch []string

// isValid returns if expression is valid
func (ex expressionMatch) isValid() bool {
	return len(ex) == groupSize
}

// variable extracts variable expression and returns validity state
func (ex expressionMatch) variable() (string, bool) {
	val := ex[varGroup]
	return val, val != ""
}

// expression extracts shell expression and returns validity state
func (ex expressionMatch) expression() (string, bool) {
	val := ex[evalGroup]
	return val, val != ""
}

// SpecV2Parser implements expression parsing for v2 spec.
type SpecV2Parser struct{}

// NewSpecV2Parser creates a new processor instance for language spec v2 for language spec v2
func NewSpecV2Parser() SpecV2Parser {
	return SpecV2Parser{}
}

// ReadString parses and evaluates expressions inside the string
func (p SpecV2Parser) ReadString(ctx EvalContext, input string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to parse string %q, %s", input, r)
		}
	}()

	var errs []error
	out := re.ReplaceAllFunc([]byte(input), func(exp []byte) []byte {
		val, err := p.ReadExpression(ctx, exp)
		if err != nil {
			errs = append(errs, fmt.Errorf("%w (at %q)", err, string(exp)))
			return exp
		}

		return val
	})

	if len(errs) > 0 {
		return "", errors.Join(errs...)
	}

	return string(out), err
}

// ContainsExpression checks if passed string contains template expressions
func (p SpecV2Parser) ContainsExpression(str string) bool {
	return re.Match([]byte(str))
}

// expandVariable expands variable value
func (p SpecV2Parser) expandVariable(ctx EvalContext, varName string) (val string, err error) {
	// trim everything for safety
	varName = strings.TrimSpace(varName)
	if varName == "" {
		return val, fmt.Errorf("expression cannot be empty")
	}

	// find the var in the scope
	val, ok := ctx.Env.ValueByName(varName)
	if !ok {
		err = fmt.Errorf("%q is not defined", varName)
		return
	}

	// Parse variable value for nested template expression
	if p.ContainsExpression(val) {
		// FIXME: check if variable refers to itself (x = {{x}})
		return p.ReadString(ctx, val)
	}

	// just return a plain value if it's not an expression
	return val, nil
}

// ReadExpression evaluates an expression string
func (p SpecV2Parser) ReadExpression(ctx EvalContext, exp []byte) (result []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()

	match := expressionMatch(re.FindStringSubmatch(string(exp)))

	if !match.isValid() {
		return exp, nil
	}

	if val, ok := match.variable(); ok {
		r, err := p.expandVariable(ctx, val)
		return []byte(r), err
	}

	if cmd, ok := match.expression(); ok {
		return p.evalExpression(ctx.CommandProcessor, cmd)
	}

	return exp, nil
}

func (p SpecV2Parser) evalExpression(c CommandProcessor, cmd string) ([]byte, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil, nil
	}

	result, err := c.EvalCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to eval inline script: %s", err)
	}

	// Trim \r\n and \n from command output
	result = []byte(strings.TrimSpace(string(result)))
	return result, err
}
