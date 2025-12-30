package golang

import (
	"strings"

	"github.com/go-gilbert/gilbert/internal/v2/manifest/argschema"
)

type commonBuildArgs struct {
	Strip      bool
	Race       bool
	CGoEnabled argschema.Option[bool]
	BuildVCS   argschema.Option[bool]

	WorkDir     string
	OS          string
	Arch        string
	BuildMode   string
	BuildTags   []string
	AsmFlags    []string
	LinkerFlags []string
	GCCGOFlags  []string
	GCFlags     []string

	PackageVars map[string]string
	Env         map[string]string
}

func (args commonBuildArgs) appendCommonArgs(parentArgs []string) []string {
	if args.Race {
		parentArgs = append(parentArgs, "-race")
	}

	if val, ok := args.BuildVCS.Value(); ok {
		if val {
			parentArgs = append(parentArgs, "-buildvcs=true")
		} else {
			parentArgs = append(parentArgs, "-buildvcs=false")
		}
	}

	if args.BuildMode != "" {
		parentArgs = append(parentArgs, "-buildmode", args.BuildMode)
	}

	if args.WorkDir != "" {
		parentArgs = append(parentArgs, "-C", args.WorkDir)
	}

	if len(args.BuildTags) > 0 {
		parentArgs = append(parentArgs, "-tags", strings.Join(args.BuildTags, ","))
	}

	ldflags := args.ldparams()
	if len(ldflags) > 0 {
		parentArgs = append(parentArgs, "-ldflags", ldflags)
	}

	if len(args.GCFlags) > 0 {
		vals := strings.Join(args.GCFlags, " ")
		parentArgs = append(parentArgs, "-gcflags", vals)
	}

	if len(args.AsmFlags) > 0 {
		vals := strings.Join(args.AsmFlags, " ")
		parentArgs = append(parentArgs, "-asmflags", vals)
	}

	if len(args.GCCGOFlags) > 0 {
		vals := strings.Join(args.GCCGOFlags, " ")
		parentArgs = append(parentArgs, "-gccgoflags", vals)
	}

	return parentArgs
}

func (args commonBuildArgs) buildEnv(parent map[string]string) []string {
	duplicates := make(map[string]struct{})
	vals := make([]string, 0, len(parent)+len(args.Env)+5)

	if args.Arch != "" {
		duplicates["GOARCH"] = struct{}{}
		vals = append(vals, "GOARCH="+args.Arch)
	}

	if args.OS != "" {
		duplicates["GOOS"] = struct{}{}
		vals = append(vals, "GOOS="+args.OS)
	}

	if val, ok := args.CGoEnabled.Value(); ok {
		duplicates["CGO_ENABLED"] = struct{}{}
		if val {
			vals = append(vals, "CGO_ENABLED=1")
		} else {
			vals = append(vals, "CGO_ENABLED=0")
		}
	}

	for k, v := range args.Env {
		if _, ok := duplicates[k]; ok {
			continue
		}

		duplicates[k] = struct{}{}
		vals = append(vals, k+"="+v)
	}

	for k, v := range parent {
		if _, ok := duplicates[k]; ok {
			continue
		}

		duplicates[k] = struct{}{}
		vals = append(vals, k+"="+v)
	}

	return vals
}

func (args commonBuildArgs) ldparams() string {
	out := &strings.Builder{}
	if args.Strip {
		out.WriteString("-s -w")
	}

	// TODO: should vals be quoted?
	for k, v := range args.PackageVars {
		if out.Len() > 0 {
			out.WriteRune(' ')
		}

		// override package vars using linker:
		// '-X Foo=Bar'
		out.WriteString("-X ")
		out.WriteString(k)
		out.WriteRune('=')
		out.WriteString(v)
	}

	for _, v := range args.LinkerFlags {
		if out.Len() > 0 {
			out.WriteRune(' ')
		}

		out.WriteString(v)
	}

	return out.String()
}
