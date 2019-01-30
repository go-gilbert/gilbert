package build

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/x1unix/guru/env"
	. "github.com/x1unix/guru/tools"
)

const (
	osEnvVar = "GOOS"
	archEnvVar = "GOARCH"
)

// LinkerParams is set of params for linker (ln)
type LinkerParams struct {
	StripDebugInfo bool
	LinkerFlags    []string
}

// CompileTarget struct contains information about compile target
type CompileTarget struct {
	Os   string
	Arch string
}

// envVars returns environment variables for platform-specific builds
func (c *CompileTarget) envVars() []string {
	return []string{
		osEnvVar + "=" + c.Os,
		archEnvVar + "=" + c.Arch,
	}
}

func newParams() Params {
	return Params{
		Target: CompileTarget{
			Os: runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}
}

// Params is params for the build plugin
type Params struct {
	Source     string
	BuildMode  string
	OutputPath string
	Params     LinkerParams
	Target     CompileTarget
	Variables  env.Vars
}

// linkerParams generates list of arguments for Go linker
func (p *Params) linkerParams(ctx *env.Context) (args []string, err error) {
	if p.Params.StripDebugInfo {
		args = append(args, "-s", "-w")
	}

	for k, v := range p.Variables {
		expanded, err := ctx.ExpandVariables(v)
		if err != nil {
			return nil, err
		}

		// override package vars using linker:
		// '-X main.Foo=Bar'
		args = append(args, "-X "+k+"="+expanded)
	}

	return append(args, p.Params.LinkerFlags...), nil
}

// buildArgs returns arguments for Go tools to build the artifact
func (p *Params) buildArgs(ctx *env.Context) (args []string, err error) {
	args = []string{"build"}

	// Add output file param
	if !StringEmpty(p.OutputPath) {
		output, err := ctx.ExpandVariables(p.OutputPath)
		if err != nil {
			return nil, err
		}

		args = append(args, "-o", output)
	}

	// Append linker params
	ldFlags, err := p.linkerParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build params for Go linker, %s", err)
	}

	if len(ldFlags) > 0 {
		args = append(args, "-ldflags", strings.Join(ldFlags, " "))
	}

	// Add build mode
	if !StringEmpty(p.BuildMode) {
		args = append(args, "-buildmode", p.BuildMode)
	}

	// Add package/file name to command
	if !StringEmpty(p.Source) {
		source, err := ctx.ExpandVariables(p.Source)
		if err != nil {
			return nil, err
		}

		args = append(args, source)
	}

	return args, nil
}

// createCompilerProcess creates compiler process to start
func (p *Params) newCompilerProcess(ctx *env.Context) (*exec.Cmd, error){
	args, err := p.buildArgs(ctx)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("go", args...)
	cmd.Env = append(ctx.Environ(), p.Target.envVars()...)
	cmd.Dir = ctx.Environment.ProjectDirectory
	return cmd, nil
}
