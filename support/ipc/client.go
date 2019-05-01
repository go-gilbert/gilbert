package ipc

import (
	"context"
	"os/exec"
	"strconv"
)

const (
	protocolVersion = 1.0
	msgPoolSz       = 10
)

func getExecArgs() []string {
	return []string{"--connect", "--protocol-version=" + strconv.Itoa(protocolVersion)}
}

type Client struct {
	*Gateway
	ctx      context.Context
	process  *exec.Cmd
	execPath string
}

func (c *Client) Connect() error {
	args := getExecArgs()
	c.process = exec.CommandContext(c.ctx, c.execPath, args...)
	c.Gateway = NewGateway(c.process.Stdout, msgPoolSz)
	c.process.Stdout = c.Gateway
	return c.process.Start()
}

func (c *Client) NewSession() *Session {
	return NewSession(c.Gateway, c.ctx)
}

// NewClient creates a new plugin IPC connection
func NewClient(ctx context.Context, execPath string) *Client {
	return &Client{
		ctx:      ctx,
		execPath: execPath,
	}
}
