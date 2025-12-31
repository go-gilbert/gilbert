package golang

import (
	. "github.com/go-gilbert/gilbert/internal/v2/manifest/argschema"
)

var commonFields = []FieldVisitor[commonBuildArgs]{
	StringField("workdir", func(dst *commonBuildArgs, val string) error {
		dst.WorkDir = val
		return nil
	}),
	StringField("os", func(dst *commonBuildArgs, val string) error {
		dst.OS = val
		return nil
	}),
	StringField("arch", func(dst *commonBuildArgs, val string) error {
		dst.Arch = val
		return nil
	}),
	StringField("buildmode", func(dst *commonBuildArgs, val string) error {
		dst.BuildMode = val
		return nil
	}),
	BoolField("buildvcs", func(dst *commonBuildArgs, val bool) error {
		dst.BuildVCS = Some(val)
		return nil
	}),
	BoolField("cgo", func(dst *commonBuildArgs, val bool) error {
		dst.CGoEnabled = Some(val)
		return nil
	}),
	BoolField("race", func(dst *commonBuildArgs, val bool) error {
		dst.Race = val
		return nil
	}),
	BoolField("strip", func(dst *commonBuildArgs, val bool) error {
		dst.Strip = val
		return nil
	}),
	ListField("tags", AnyString, func(dst *commonBuildArgs, val []string) error {
		dst.BuildTags = val
		return nil
	}),
	ListField("gccgoflags", AnyString, func(dst *commonBuildArgs, val []string) error {
		dst.GCCGOFlags = val
		return nil
	}),
	ListField("gcflags", AnyString, func(dst *commonBuildArgs, val []string) error {
		dst.GCFlags = val
		return nil
	}),
	ListField("ldflags", AnyString, func(dst *commonBuildArgs, val []string) error {
		dst.LinkerFlags = val
		return nil
	}),
	ListField("asmflags", AnyString, func(dst *commonBuildArgs, val []string) error {
		dst.AsmFlags = val
		return nil
	}),
	DictField("package_vars", AnyString, func(dst *commonBuildArgs, val map[string]string) error {
		dst.PackageVars = val
		return nil
	}),
	DictField("env", AnyString, func(dst *commonBuildArgs, val map[string]string) error {
		dst.Env = val
		return nil
	}),
}
