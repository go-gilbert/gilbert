package golang

import (
	"errors"
	"strings"

	. "github.com/go-gilbert/gilbert/internal/manifest/argschema"
)

var buildSchema = Struct(
	StringField("package", func(dst *buildActionArgs, val string) error {
		val = strings.TrimSpace(val)
		if val == "" {
			return errors.New("package name is required")
		}

		dst.PackageName = val
		return nil
	}).Required(),
	StringField("output", func(dst *buildActionArgs, val string) error {
		dst.Output = val
		return nil
	}),
).
	Embedded(Embedded(func(parent *buildActionArgs) *commonBuildArgs {
		return &parent.commonBuildArgs
	}, commonFields...))

type buildActionArgs struct {
	commonBuildArgs
	PackageName string
	Output      string
}

func (args buildActionArgs) commandArgs() ([]string, error) {
	if args.PackageName == "" {
		return nil, errors.New("missing package name")
	}

	argv := []string{"build"}
	if args.Output != "" {
		argv = append(argv, "-o", args.Output)
	}

	argv = args.commonBuildArgs.appendCommonArgs(argv)
	argv = append(argv, args.PackageName)
	return argv, nil
}

// TODO: validation
