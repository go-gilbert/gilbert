package os

import (
	"errors"
	"strings"

	. "github.com/go-gilbert/gilbert/internal/manifest/argschema"
	"github.com/go-gilbert/gilbert/pkg/executil"
)

var shellSchema = Struct(
	StringField("command", func(dst *shellActionArgs, val string) error {
		val = strings.TrimSpace(val)
		if val == "" {
			return errors.New("command is required")
		}

		dst.Command = val
		return nil
	}).Required(),
	StringField("workdir", func(dst *shellActionArgs, val string) error {
		dst.WorkDir = val
		return nil
	}),
	BoolField("silent", func(dst *shellActionArgs, val bool) error {
		dst.Silent = val
		return nil
	}),
	StringField("shell", func(dst *shellActionArgs, val string) error {
		dst.Shell = val
		return nil
	}),
	StringField("shellCommand", func(dst *shellActionArgs, val string) error {
		// TODO: move into struct
		dst.ShellCommand = val
		return nil
	}),
	DictField("env", AnyString, func(dst *shellActionArgs, val map[string]string) error {
		dst.Env = val
		return nil
	}),
).Constructor(func(args *shellActionArgs) {
	args.Shell = executil.ShellPath
	args.ShellCommand = executil.ShellCmdOption
})

type shellActionArgs struct {
	Silent       bool
	Command      string
	WorkDir      string
	Shell        string
	ShellCommand string
	Env          map[string]string
}
