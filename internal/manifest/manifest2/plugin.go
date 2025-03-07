package manifest2

import (
	"errors"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
)

var urlSchemeRe = regexp.MustCompile(`(?m)^[a-z0-9\+\.\-]+://`)

// PluginImports is a key-value pair of namespace and plugin import.
type PluginImports = map[string]*PluginImport

type PluginImport struct {
	// Location specifies source declaration of import statement.
	Location *ReferenceLocation

	// URI is plugin import URI.
	URI string
}

// PathIntoURI checks whether path is an URL or file path.
//
// If path is a file path - makes path absolute based on working directory
// and returns `file://` URL.
func PathIntoURI(p, workDir string) (string, error) {
	if p == "" {
		return "", errors.New("empty path")
	}

	if urlSchemeRe.MatchString(p) {
		return urlIntoString(url.Parse(p))
	}

	if runtime.GOOS == "windows" {
		p = filepath.ToSlash(p)
	}

	p = filepath.Join(workDir, p)
	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}

	return urlIntoString(url.ParseRequestURI("file://" + absPath))
}

func urlIntoString(u *url.URL, err error) (string, error) {
	if err != nil {
		return "", err
	}

	return u.String(), nil
}
