package os

import (
	"errors"
	"strings"

	. "github.com/go-gilbert/gilbert/internal/v2/manifest/argschema"
)

var processSchema = Struct(
	StringField("path", func(dst *processActionArgs, val string) error {
		val = strings.TrimSpace(val)
		if val == "" {
			return errors.New("process path is required")
		}

		dst.Path = val
		return nil
	}).Required(),
	ListField("args", AnyString, func(dst *processActionArgs, val []string) error {
		dst.Args = val
		return nil
	}),
	StringField("workdir", func(dst *processActionArgs, val string) error {
		dst.WorkDir = val
		return nil
	}),
	BoolField("silent", func(dst *processActionArgs, val bool) error {
		dst.Silent = val
		return nil
	}),
	DictField("env", AnyString, func(dst *processActionArgs, val map[string]string) error {
		dst.Env = val
		return nil
	}),
)

type processActionArgs struct {
	Silent  bool
	Path    string
	WorkDir string
	Args    []string
	Env     map[string]string
}
