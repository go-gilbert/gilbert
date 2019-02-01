package scope

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// FIXME: regex fails to capture shell expressions with "%" and "}" chars (might affect winnt envs).

var re = regexp.MustCompile(templateRegEx)

const (
	evalGroup = 1 // match group that contains shell eval expression
	varGroup  = 2 // match group that contains variable print expression

	groupSize = 3 // match group size

	// Regular expression to extract template expressions
	//
	// - Shell-eval string:
	//		{% whoami %}
	// - Variable value
	//		{{ foo }}
	//
	templateRegEx = `(?m){%([^%}]+)%}|{{([\s\d\w]+)}}`
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

// ExpressionProcessor evaluates template literals inside the string
type ExpressionProcessor struct {
	ctx *Context
}

// NewExpressionProcessor creates a new processor instance
func NewExpressionProcessor(ctx *Context) ExpressionProcessor {
	return ExpressionProcessor{ctx: ctx}
}

// ReadString parses and evaluates expressions inside the string
func (p *ExpressionProcessor) ReadString(input string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to parse string '%s', %s", input, r)
		}
	}()

	out := re.ReplaceAllFunc([]byte(input), func(exp []byte) []byte {
		val, err := p.ReadExpression(exp)
		if err != nil {
			// TODO: find a better way to pass error from the callback
			panic(fmt.Errorf("%s (at '%s')", err, string(exp)))
		}

		return val
	})

	return string(out), err
}

// ContainsExpression checks if passed string contains template expressions
func (p *ExpressionProcessor) ContainsExpression(str string) bool {
	return re.Match([]byte(str))
}

// expandVariable expands variable value
func (p *ExpressionProcessor) expandVariable(varName string) (val string, err error) {
	if p.ctx == nil {
		panic("processor context is undefined")
	}

	// trim everything for safety
	varName = strings.TrimSpace(varName)
	if varName == "" {
		return val, fmt.Errorf("expression cannot be empty")
	}

	// find the var in the scope
	_, val, ok := p.ctx.Var(varName)
	if !ok {
		err = fmt.Errorf("variable '%s' is undefined", varName)
		return
	}

	// Parse variable value for nested template expression
	if p.ContainsExpression(val) {
		// FIXME: check if variable refers to itself (x = {{x}})
		return p.ReadString(val)
	}

	// just return a plain value if it's not an expression
	return val, nil
}

// ReadExpression evaluates an expression string
func (p *ExpressionProcessor) ReadExpression(exp []byte) (result []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occured: %v", r)
		}
	}()

	match := expressionMatch(re.FindStringSubmatch(string(exp)))

	if !match.isValid() {
		return exp, nil
	}

	if val, ok := match.variable(); ok {
		r, err := p.expandVariable(val)
		return []byte(r), err
	}

	if _, ok := match.expression(); ok {
		return nil, errors.New("shell execution expression is not supported now")
	}

	return exp, nil
}
