package docker

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	sdk "github.com/go-gilbert/gilbert-sdk"
)

type ContainerCreateAction struct {
	client *client.Client
	args   containerArgs
	scope  sdk.ScopeAccessor
}

func (a *ContainerCreateAction) buildImage(ctx sdk.JobContextAccessor) error {
	opts := types.ImageBuildOptions{}
	body := &bytes.Reader{}

	buildPath := a.args.Build
	if !filepath.IsAbs(buildPath) {
		env := a.scope.Environment()
		buildPath = filepath.Join(env.ProjectDirectory, buildPath)
	}

	ctx.Log().Infof("Building image from %q ...", buildPath)
	opts.Tags = []string{filepath.Base(buildPath)}
	opts.Dockerfile = buildPath
	opts.PullParent = true

	_, err := a.client.ImageBuild(ctx.Context(), body, opts)
	if err != nil {
		return fmt.Errorf("failed to build container image: %s", err)
	}

	return nil
}

func (a *ContainerCreateAction) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) error {
	if a.args.CanBeBuilt() {
		if err := a.buildImage(ctx); err != nil {
			return err
		}
	} else {
		ctx.Log().Infof("Pulling %q...", a.args.Image)
		_, err := a.client.ImagePull(ctx.Context(), a.args.Image, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image %q: %s", a.args.Image, err)
		}
	}

	return nil
}

func (a *ContainerCreateAction) Cancel(ctx sdk.JobContextAccessor) error {
	if err := a.client.Close(); err != nil {
		ctx.Log().Debugf("docker: failed to close client connection - %s", err)
	}

	return nil
}

// NewContainerCreateAction creates a new "docker:create-container" action handler
func NewContainerCreateAction(s sdk.ScopeAccessor, a sdk.ActionParams) (sdk.ActionHandler, error) {
	args := containerArgs{}
	if err := a.Unmarshal(&args); err != nil {
		return nil, err
	}

	if err := args.Validate(); err != nil {
		return nil, err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %s", err)
	}

	return &ContainerCreateAction{client: cli, args: args, scope: s}, nil
}
