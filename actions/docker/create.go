package docker

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	sdk "github.com/go-gilbert/gilbert-sdk"
)

type ContainerCreateAction struct {
	client    *client.Client
	container containerArgs
	scope     sdk.ScopeAccessor
}

func (a *ContainerCreateAction) buildImage(ctx sdk.JobContextAccessor) (string, error) {
	opts := types.ImageBuildOptions{}
	body := &bytes.Reader{}

	buildPath := a.container.Build
	if !filepath.IsAbs(buildPath) {
		env := a.scope.Environment()
		buildPath = filepath.Join(env.ProjectDirectory, buildPath)
	}

	ctx.Log().Infof("Building image from %q ...", buildPath)
	imgName := filepath.Base(buildPath)
	opts.Tags = []string{imgName}
	opts.Dockerfile = buildPath
	opts.PullParent = true

	_, err := a.client.ImageBuild(ctx.Context(), body, opts)
	if err != nil {
		return imgName, fmt.Errorf("failed to build container image: %s", err)
	}

	return imgName, nil
}

func (a *ContainerCreateAction) ensureImage(ctx sdk.JobContextAccessor) (string, error) {
	if a.container.CanBeBuilt() {
		return a.buildImage(ctx)
	}

	ctx.Log().Infof("Pulling %q...", a.container.Image)
	_, err := a.client.ImagePull(ctx.Context(), a.container.Image, types.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image %q: %s", a.container.Image, err)
	}

	return a.container.Image, nil
}

func (a *ContainerCreateAction) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) error {
	imgName, err := a.ensureImage(ctx)
	if err != nil {
		return err
	}

	// TODO: add values from passed action params
	containerCfg := &container.Config{
		Image: imgName,
	}
	hostCfg := a.container.HostConfig()
	netCfg := a.container.NetworkingConfig()
	resp, err := a.client.ContainerCreate(ctx.Context(), containerCfg, hostCfg, netCfg, a.container.Name)
	if err != nil {
		return fmt.Errorf("failed to create container %q: %s", a.container.Name, err)
	}

	// Print warnings from response (if any)
	if len(resp.Warnings) > 0 {
		for _, w := range resp.Warnings {
			ctx.Log().Warn(w)
		}
	}

	ctx.Log().Successf("Created container %q (ID: %q)", a.container.Name, resp.ID)
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

	return &ContainerCreateAction{client: cli, container: args, scope: s}, nil
}
