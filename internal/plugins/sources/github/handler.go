package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/plugins/support"
	"github.com/go-gilbert/gilbert/internal/storage"
	"github.com/go-gilbert/gilbert/internal/support/fs"
	"github.com/go-gilbert/gilbert/internal/support/web"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v25/github"
)

// GetPlugin retrieves plugin from GitHub
func GetPlugin(ctx context.Context, uri *url.URL) (string, error) {
	dc, err := readURL(ctx, uri)
	if err != nil {
		return "", err
	}

	dir, err := storage.Path(storage.Plugins, dc.pkg.directory())
	if err != nil {
		return "", err
	}

	pluginPath := filepath.Join(dir, dc.pkg.fileName())
	exists, err := fs.Exists(pluginPath)
	if err != nil {
		return "", err
	}

	if !exists {
		log.Default.Debugf("github: plugin is not cached and need to be downloaded")
		log.Default.Debugf("github: init plugin directory: '%s'", dir)
		if err = os.MkdirAll(dir, support.PluginPermissions); err != nil {
			return "", err
		}

		asset, err := getPluginRelease(ctx, dc.ghClient, dc.pkg)
		if err != nil {
			return "", err
		}

		assetURL := asset.GetBrowserDownloadURL()
		if assetURL == "" {
			return "", errors.New("missing asset download URL")
		}

		log.Default.Debugf("github: downloading plugin from '%s'...", assetURL)
		if err := web.ProgressDownloadFile(dc.httpClient, assetURL, pluginPath); err != nil {
			return "", err
		}
	}

	return pluginPath, nil
}

func getPluginRelease(ctx context.Context, client *github.Client, pkg packageQuery) (asset *github.ReleaseAsset, err error) {
	log.Default.Logf("Downloading plugin from GitHub repo '%s/%s'", pkg.owner, pkg.repo)
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
