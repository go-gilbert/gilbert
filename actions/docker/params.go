package docker

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type containerArgs struct {
	Force       bool              `mapstructure:"force"`
	Name        string            `mapstructure:"container_name"`
	Image       string            `mapstructure:"image"`
	Build       string            `mapstructure:"build"`
	Ports       []string          `mapstructure:"ports"`
	Environment map[string]string `mapstructure:"environment"`
	Volumes     []string          `mapstructure:"volumes"`
	Expose      []string          `mapstructure:"expose"`
	Links       []string          `mapstructure:"links"`
}

func (a *containerArgs) CanBeBuilt() bool {
	return a.Build != ""
}

func (a *containerArgs) HostConfig() *container.HostConfig {
	// TODO: add implementation
	return nil
}

func (a *containerArgs) NetworkingConfig() *network.NetworkingConfig {
	// TODO: add implementation
	return nil
}

//func (a *containerArgs) PortSet() []nat.Port {
//	if len(a.Ports) == 0 {
//		return nil
//	}
//
//	for _, portSpec := range a.Ports {
//
//	}
//}

func (a *containerArgs) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("container name should not be empty. Please specify 'container_name' param")
	}

	if a.Image == "" && a.Build == "" {
		return fmt.Errorf("container params should contain base image or custom Dockerfile path. Please specify 'image' or 'build' param")
	}

	return nil
}
