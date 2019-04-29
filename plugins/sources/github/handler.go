package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-gilbert/gilbert/storage"

	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/plugins/loader"
	"github.com/go-gilbert/gilbert/tools/fs"
	"github.com/go-gilbert/gilbert/tools/web"
	"github.com/google/go-github/v25/github"
)

// ImportHandler is GitHub source import handler for plugins
func ImportHandler(ctx context.Context, uri *url.URL) (sdk.PluginFactory, string, error) {
	dc, err := readUrl(ctx, uri)
	if err != nil {
		return nil, "", err
	}

	dir, err := storage.Path(storage.Plugins, dc.pkg.directory())
	if err != nil {
		return nil, "", err
	}

	pluginPath := filepath.Join(dir, dc.pkg.fileName())
	exists, err := fs.Exists(pluginPath)
	if err != nil {
		return nil, "", err
	}

	if !exists {
		log.Default.Debugf("github: init plugin directory: '%s'", dir)
		if err = os.MkdirAll(dir, 0644); err != nil {
			return nil, "", err
		}

		asset, err := getPluginRelease(ctx, dc.ghClient, dc.pkg)
		if err != nil {
			return nil, "", err
		}

		assetUrl := asset.GetBrowserDownloadURL()
		if assetUrl == "" {
			return nil, "", errors.New("missing asset download URL")
		}

		log.Default.Debugf("github: downloading plugin from '%s'...", assetUrl)
		if err := web.ProgressDownloadFile(dc.httpClient, assetUrl, pluginPath); err != nil {
			return nil, "", err
		}
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
	log.Default.Debugf("github: trying to find release asset '%s'", assetName)
	return findReleaseAsset(assetName, rel.Assets)
}

func findReleaseAsset(fileName string, assets []github.ReleaseAsset) (*github.ReleaseAsset, error) {
	for _, asset := range assets {
		if asset.GetName() == fileName {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("repository does not contain release for platform %s %s", runtime.GOOS, runtime.GOARCH)
}
