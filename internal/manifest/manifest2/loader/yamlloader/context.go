package yamlloader

import (
	"context"
	"errors"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	"github.com/hashicorp/go-set/v3"
)

type fileInfoCtxKeyType struct{}

var fileInfoCtxKey = fileInfoCtxKeyType{}

type loaderContext struct {
	filePath        string
	fileDir         string
	dst             *manifest2.JobFile
	knownNamespaces *set.Set[string]
}

func newLoaderContext(parentCtx context.Context, info loaderContext) context.Context {
	return context.WithValue(parentCtx, fileInfoCtxKey, &info)
}

func getLoaderContext(ctx context.Context) (*loaderContext, error) {
	v := ctx.Value(fileInfoCtxKey)
	if v == nil {
		return nil, errors.New("context is not loader context")
	}

	c, ok := v.(*loaderContext)
	if !ok {
		return nil, errors.New("invalid context key value")
	}

	return c, nil
}

type jobGroupCtxKeyType struct{}

var jobGroupCtxKey = jobGroupCtxKeyType{}

type jobGroupInfo struct {
	jobGroupType manifest2.JobGroupType
}

func jobGroupContext(parentCtx context.Context, info *jobGroupInfo) context.Context {
	return context.WithValue(parentCtx, jobGroupCtxKey, info)
}

func jobGroupInfoFromContext(ctx context.Context) (*jobGroupInfo, bool) {
	v := ctx.Value(jobGroupCtxKey)
	if v == nil {
		return nil, false
	}

	c, ok := v.(*jobGroupInfo)
	return c, ok
}
