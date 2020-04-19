package runner

import (
	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/v2/manifest"
	"github.com/hashicorp/hcl/v2"
)

type TaskRunner struct {
	log      sdk.Logger
	manifest *manifest.Manifest
}

func NewTaskRunner(log sdk.Logger, m *manifest.Manifest) TaskRunner {
	return TaskRunner{
		log:      log,
		manifest: m,
	}
}

func (tr TaskRunner) RunTask(t *manifest.Task, ctx *hcl.EvalContext) error {
	return nil
}
