package expr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"unsafe"

	"github.com/expr-lang/expr/checker"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/file"
	"github.com/expr-lang/expr/optimizer"
	"github.com/expr-lang/expr/parser"
)

func valueToBytes(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	switch t := v.(type) {
	case string:
		return []byte(t), nil
	case []byte:
		return t, nil
	case bool, uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
		buff := bytes.NewBuffer(make([]byte, 0, 10))
		_, err := fmt.Fprint(buff, v)
		return buff.Bytes(), err
	case fmt.Stringer:
		return []byte(t.String()), nil
	}

	return nil, fmt.Errorf("value of type %T cannot be converted to a string", v)
}

func bytesAsString(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	return unsafe.String(unsafe.SliceData(b), len(b))
}

func evalExprFromBody(c *conf.Config, body Expression) (*parser.Tree, error) {
	// Lexer already bans dynamic expressions inside eval expression.
	// Just double-check for just in case.
	if body.Evaluable() {
		return nil, errors.New("expressions are not allowed in this context")
	}

	src, err := body.EvalText(context.TODO(), NoopEvalParams)
	if err != nil {
		return nil, err
	}

	return parseEvalExpr(c, bytesAsString(src))
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
