package expr

import (
	"errors"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/checker"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/file"
	"github.com/expr-lang/expr/optimizer"
	"github.com/expr-lang/expr/parser"
)

func valueToString(v any) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	case bool, uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
		return fmt.Sprint(v), nil
	}

	return "", fmt.Errorf("value %#v cannot be converted to a string", v)
}

func evalConfWithOptions(opts ...expr.Option) *conf.Config {
	c := conf.CreateNew()
	for _, op := range opts {
		op(c)
	}
	for name := range c.Disabled {
		delete(c.Builtins, name)
	}
	c.Check()

	return c
}

func parseEvalExpr(c *conf.Config, str string) (*parser.Tree, error) {
	tree, err := checker.ParseCheck(str, c)
	if err != nil {
		return nil, err
	}

	if c.Optimize {
		err = optimizer.Optimize(&tree.Node, c)
		if err != nil {
			var fileError *file.Error
			if errors.As(err, &fileError) {
				err = fileError.Bind(tree.Source)
			}

			return nil, err
		}
	}

	return tree, nil
}
