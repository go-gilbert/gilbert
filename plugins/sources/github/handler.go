package github

import (
	"context"
	"github.com/go-gilbert/gilbert-sdk"
	"net/url"
)

func ImportHandler(ctx context.Context, uri *url.URL) (sdk.PluginFactory, string, error) {
	_, _, err := readUrl(ctx, uri)
	if err != nil {
		return nil, "", err
	}

	return nil, "", nil
}
