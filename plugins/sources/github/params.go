package github

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v25/github"
)

const (
	defaultDomain   = "github.com"
	defaultProtocol = "https"
	pkgPathSize     = 2 // /owner/repo
	pathDelimiter   = "/"
	protocolParam   = "protocol"
	versionParam    = "version"
	tokenParam      = "token"
)

var (
	errNoHost = errors.New("please specify github host (e.g github.com)")
	errBadUrl = errors.New("bad GitHub repo path format (expected: 'github.com/owner/repo_name')")
)

type packageQuery struct {
	owner   string
	repo    string
	version string
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

func readUrl(ctx context.Context, uri *url.URL) (client *github.Client, pkg packageQuery, err error) {
	if uri.Host == "" {
		return nil, pkg, errNoHost
	}

	// Trim slashes around package path and split path
	pkgPath := strings.Split(strings.Trim(uri.Path, pathDelimiter), pathDelimiter)
	if len(pkgPath) < pkgPathSize {
		return nil, pkg, errBadUrl
	}

	httpClient := getHttpClient(ctx, uri)
	if uri.Host != defaultDomain {
		// Handle enterprise url if used non-default domain
		var ghUrl string
		ghUrl, pkg = parseEnterpriseUrl(uri, pkgPath)

		client, err = github.NewEnterpriseClient(ghUrl, ghUrl, httpClient)
	} else {
		pkg = parsePkgPath(pkgPath)
		client = github.NewClient(httpClient)
	}

	pkg.version = uri.Query().Get(versionParam)
	return client, pkg, err
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
