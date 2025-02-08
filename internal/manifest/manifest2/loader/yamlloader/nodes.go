package yamlloader

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/goccy/go-yaml/ast"
)

func readVersionNode(n ast.Node) (string, error) {
	if n == nil {
		return "", errors.New("version field is required")
	}

	version := ""
	switch t := n.(type) {
	case *ast.IntegerNode:
		version = strconv.Itoa(t.Value.(int))
	case *ast.FloatNode:
		version = t.Token.Value
	case *ast.StringNode:
		version = t.Value
	default:
		return "", errors.New("version node should be string value")
	}

	if version == "2" {
		return "", fmt.Errorf("unsupported version: %q", version)
	}

	return version, nil
}

func readStringNode(n ast.Node) (string, error) {
	if s, ok := n.(*ast.StringNode); ok {
		return s.Value, nil
	}

	return "", errors.New("value should be string")
}
