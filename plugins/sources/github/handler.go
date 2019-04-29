package github

import (
	"context"
	"fmt"
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/plugins/loader"
	"github.com/go-gilbert/gilbert/tools/fs"
	"github.com/google/go-github/v25/github"
	"net/url"
	"path/filepath"
	"runtime"
)

func ImportHandler(ctx context.Context, uri *url.URL) (sdk.PluginFactory, string, error) {
	client, pkg, err := readUrl(ctx, uri)
	if err != nil {
		return nil, "", err
	}

	dir, err := pkg.directory()
	if err != nil {
		return nil, "", err
	}

	pluginPath := filepath.Join(dir, pkg.fileName())
	exists, err := fs.Exists(pluginPath)
	if err != nil {
		return nil, "", err
	}

	if !exists {
		// TODO: download library
	}

	return loader.LoadLibrary(pluginPath)
}

func getPluginRelease(ctx context.Context, client *github.Client, pkg packageQuery) (asset *github.ReleaseAsset, err error) {
	log.Default.Infof("Downloading release '%s' from '@%s/%s'", pkg.owner, pkg.repo)
	var rel *github.RepositoryRelease
	if pkg.version == latestVersion {
		rel, _, err = client.Repositories.GetLatestRelease(ctx, pkg.owner, pkg.repo)
	} else {
		rel, _, err = client.Repositories.GetReleaseByTag(ctx, pkg.owner, pkg.repo, pkg.version)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get release information from GitHub, %s", err)
	}

	assetName := pkg.fileName()
	log.Default.Debugf("Trying to find release asset '%s'", assetName)
	return findReleaseAsset(assetName, rel.Assets)
}

//func downloadPlugin(ctx context.Context, asset *github.ReleaseAsset, dest string) error {
//	url := asset.GetBrowserDownloadURL()
//	log.Default.Debugf("Downloading plugin from '%s'...", url)
//
//}

func findReleaseAsset(fileName string, assets []github.ReleaseAsset) (*github.ReleaseAsset, error) {
	for _, asset := range assets {
		if asset.GetName() == fileName {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("repository does not contain release for platform %s %s", runtime.GOOS, runtime.GOARCH)
}
