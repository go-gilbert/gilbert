package yamlloader

import (
	"context"
	"errors"
)

type fileInfoCtxKeyType struct{}

var fileInfoCtxKey = fileInfoCtxKeyType{}

type loaderContext struct {
	fileDir string
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
