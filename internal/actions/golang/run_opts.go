package golang

import (
	"errors"
	"strings"

	. "github.com/go-gilbert/gilbert/internal/manifest/argschema"
)

var runSchema = Struct(
	StringField("package", func(dst *runActionArgs, val string) error {
		val = strings.TrimSpace(val)
		if val == "" {
			return errors.New("package name is required")
		}

		dst.PackageName = val
		return nil
	}).Required(),
	ListField("output", AnyString, func(dst *runActionArgs, val []string) error {
		dst.Args = val
		return nil
	}),
).
	Embedded(Embedded(func(parent *runActionArgs) *commonBuildArgs {
		return &parent.commonBuildArgs
	}, commonFields...))

type runActionArgs struct {
	commonBuildArgs
	PackageName string
	Args        []string
}

func (args runActionArgs) commandArgs() ([]string, error) {
	if args.PackageName == "" {
		return nil, errors.New("missing package name")
	}

	argv := []string{"run"}
	argv = args.appendCommonArgs(argv)
	argv = append(argv, args.PackageName)
	argv = append(argv, args.Args...)
	return argv, nil
}
