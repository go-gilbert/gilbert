package golang

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert/internal/v2/engine"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/argschema"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/executil"
)

type buildActionArgs struct {
	CGoEnabled bool
	BuildVCS   bool
	Strip      bool
	Race       bool

	PackageName string
	Output      string
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

func (args buildActionArgs) commandArgs() ([]string, error) {
	if args.PackageName == "" {
		return nil, errors.New("missing package name")
	}

	argv := []string{"build"}
	if args.Race {
		argv = append(argv, "-race")
	}

	if !args.BuildVCS {
		argv = append(argv, "-buildvcs=false")
	}

	if args.BuildMode != "" {
		argv = append(argv, "-buildmode", args.BuildMode)
	}

	if args.WorkDir != "" {
		argv = append(argv, "-C", args.WorkDir)
	}

	if args.Output != "" {
		argv = append(argv, "-o", args.Output)
	}

	if len(args.BuildTags) > 0 {
		argv = append(argv, "-tags", strings.Join(args.BuildTags, ","))
	}

	ldflags := args.ldparams()
	if len(ldflags) > 0 {
		argv = append(argv, "-ldflags", ldflags)
	}

	if len(args.GCFlags) > 0 {
		vals := strings.Join(args.GCFlags, " ")
		argv = append(argv, "-gcflags", vals)
	}

	if len(args.AsmFlags) > 0 {
		vals := strings.Join(args.AsmFlags, " ")
		argv = append(argv, "-asmflags", vals)
	}

	if len(args.GCCGOFlags) > 0 {
		vals := strings.Join(args.GCCGOFlags, " ")
		argv = append(argv, "-gccgoflags", vals)
	}

	argv = append(argv, args.PackageName)
	return argv, nil
}

func (args buildActionArgs) buildEnv(parent map[string]string) []string {
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

func (args buildActionArgs) ldparams() string {
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

// TODO: validation
var _ engine.ActionHandler = (*BuildActionHandler)(nil)

type BuildActionHandler struct {
	logger log.Logger
	shell  *engine.Shell

	globals scope.Globals
	args    buildActionArgs
}

func NewBuildActionHandler(ctx context.Context, ref manifest.JobHandlerRef, params engine.ActionParams) engine.HandlerResult {
	args, diags := argschema.MapArgsToStruct(ctx, params, buildSchema)
	if diags.HasError() {
		return engine.NewBadActionParamsResult(ref, diags)
	}

	h := &BuildActionHandler{
		logger:  params.Logger,
		shell:   params.Shell,
		globals: params.Scope.Globals,
		args:    args,
	}

	return engine.NewHandlerResult(h, diags)
}

func (b *BuildActionHandler) HandleAction(ctx context.Context) error {
	argv, err := b.args.commandArgs()
	if err != nil {
		return err
	}

	env := b.args.buildEnv(b.globals.Env)
	cmd := exec.CommandContext(ctx, "go", argv...)
	cmd.Env = env
	cmd.Dir = b.globals.Project.WorkDir

	w := b.logger.IOWriter()
	cmd.Stdout = w.Stdout
	cmd.Stderr = w.Stderr

	b.logger.Debugw(
		"starting command",
		log.NewField("cwd", cmd.Dir),
		log.NewField("cmd", cmd.Args),
		log.NewField("env", cmd.Env),
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start go command: %w", err)
	}

	b.logger.Infof("building %s", b.args.PackageName)

	if err := cmd.Wait(); err != nil {
		return executil.FormatExitError(err)
	}

	return errors.New("not implemented")
}
