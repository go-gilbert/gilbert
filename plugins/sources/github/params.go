package github

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-gilbert/gilbert/plugins/support"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v25/github"
)

const (
	defaultDomain   = "github.com"
	latestVersion   = "latest"
	defaultProtocol = "https"
	pkgPathSize     = 2 // /owner/repo
	pathDelimiter   = "/"
	protocolParam   = "protocol"
	versionParam    = "version"
	tokenParam      = "token"

	// ProviderName is GitHub provider name
	ProviderName = "github"
)

var (
	errNoHost = errors.New("please specify github host (e.g github.com)")
	errBadUrl = errors.New("bad GitHub repo path format (expected: 'github.com/owner/repo_name')")
)

type downloadContext struct {
	ghClient   *github.Client
	httpClient *http.Client
	pkg        packageQuery
}

type packageQuery struct {
	owner    string
	repo     string
	version  string
	location string
}

func (p *packageQuery) fileName() string {
	name := fmt.Sprintf("%s_%s-%s", p.repo, runtime.GOOS, runtime.GOARCH)
	return support.AddPluginExtension(name)
}

func (p *packageQuery) directory() string {
	hasher := md5.New()
	hasher.Write([]byte(p.location))
	return filepath.Join(providerName, hex.EncodeToString(hasher.Sum(nil)))
}

func getHttpClient(ctx context.Context, uri *url.URL) *http.Client {
	if token := uri.Query().Get(tokenParam); token != "" {
		// use oauth2 client if access token presents
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token,
		})

		return oauth2.NewClient(ctx, ts)
	}

	return &http.Client{}
}

func readUrl(ctx context.Context, uri *url.URL) (*downloadContext, error) {
	if uri.Host == "" {
		return nil, errNoHost
	}

	// Trim slashes around package path and split path
	pkgPath := strings.Split(strings.Trim(uri.Path, pathDelimiter), pathDelimiter)
	if len(pkgPath) < pkgPathSize {
		return nil, errBadUrl
	}

	dc := downloadContext{
		httpClient: getHttpClient(ctx, uri),
	}

	var err error
	if uri.Host != defaultDomain {
		// Handle enterprise url if used non-default domain
		var ghUrl string
		ghUrl, dc.pkg = parseEnterpriseUrl(uri, pkgPath)
		dc.ghClient, err = github.NewEnterpriseClient(ghUrl, ghUrl, dc.httpClient)
	} else {
		dc.pkg = parsePkgPath(pkgPath)
		dc.ghClient = github.NewClient(dc.httpClient)
	}

	if ver := uri.Query().Get(versionParam); ver != "" {
		dc.pkg.version = ver
	} else {
		dc.pkg.version = latestVersion
	}

	dc.pkg.location = path.Join(uri.Hostname(), uri.Path, dc.pkg.version)
	return &dc, err
}

func parsePkgPath(pkgPath []string) packageQuery {
	return packageQuery{
		owner: pkgPath[0],
		repo:  pkgPath[1],
	}
}

func parseEnterpriseUrl(uri *url.URL, urlPath []string) (out string, pkg packageQuery) {
	// Determine protocol
	if proto := uri.Query().Get(protocolParam); proto != "" {
		out += proto
	} else {
		out += defaultProtocol
	}

	// Determine github enterprise URL
	out += "://" + uri.Host

	// extract additional path if present
	// path presents before package path (e.g. github.example.com/custom_path/owner/repo
	if urlPathLen := len(urlPath); urlPathLen > pkgPathSize {
		out += pathDelimiter + strings.Join(urlPath[:urlPathLen-pkgPathSize], pathDelimiter)
		pkgPath := urlPath[urlPathLen-pkgPathSize:]

		pkg.owner = pkgPath[0]
		pkg.repo = pkgPath[1]
	} else {
		pkg.owner = urlPath[0]
		pkg.repo = urlPath[1]
	}

	return out, pkg
}
