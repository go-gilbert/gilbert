package golang

import (
	"errors"
	"strings"

	. "github.com/go-gilbert/gilbert/internal/v2/manifest/argschema"
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
	StringField("workdir", func(dst *buildActionArgs, val string) error {
		dst.WorkDir = val
		return nil
	}),
	StringField("os", func(dst *buildActionArgs, val string) error {
		dst.OS = val
		return nil
	}),
	StringField("arch", func(dst *buildActionArgs, val string) error {
		dst.Arch = val
		return nil
	}),
	StringField("buildmode", func(dst *buildActionArgs, val string) error {
		dst.BuildMode = val
		return nil
	}),
	BoolField("buildvcs", func(dst *buildActionArgs, val bool) error {
		dst.BuildVCS = val
		return nil
	}),
	BoolField("cgo", func(dst *buildActionArgs, val bool) error {
		dst.CGoEnabled = val
		return nil
	}),
	BoolField("race", func(dst *buildActionArgs, val bool) error {
		dst.Race = val
		return nil
	}),
	BoolField("strip", func(dst *buildActionArgs, val bool) error {
		dst.Strip = val
		return nil
	}),
	ListField("tags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.BuildTags = val
		return nil
	}),
	ListField("gccgoflags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.GCCGOFlags = val
		return nil
	}),
	ListField("gcflags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.GCFlags = val
		return nil
	}),
	ListField("ldflags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.LinkerFlags = val
		return nil
	}),
	ListField("asmflags", AnyString, func(dst *buildActionArgs, val []string) error {
		dst.AsmFlags = val
		return nil
	}),
	Field("package_vars", Dict(AnyString), func(_ VisitParams, dst *buildActionArgs, val map[string]string) error {
		dst.PackageVars = val
		return nil
	}),
	Field("env", Dict(AnyString), func(_ VisitParams, dst *buildActionArgs, val map[string]string) error {
		dst.Env = val
		return nil
	}),
).Constructor(func(v *buildActionArgs) {
	// On by default on go
	v.BuildVCS = true
})
