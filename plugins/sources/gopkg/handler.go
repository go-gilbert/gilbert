package gopkg

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/plugins/support"
	"github.com/go-gilbert/gilbert/storage"
	"github.com/go-gilbert/gilbert/support/fs"
	"net/url"
	"os/exec"
	"path/filepath"
)

type importContext struct {
	pkgPath  string
	fileName string
}

func (i *importContext) pluginPath() (string, error) {
	hasher := md5.New()
	hasher.Write([]byte(i.pkgPath))
	pluginDir, err := storage.LocalPath(storage.Plugins, hex.EncodeToString(hasher.Sum(nil)))
	if err != nil {
		return pluginDir, err
	}

	return filepath.Join(pluginDir, i.fileName), nil
}

func newImportContext(uri *url.URL) *importContext {
	pkgPath := filepath.Clean(filepath.Join(uri.Host + "/" + uri.Path))
	fName := support.AddPluginExtension(filepath.Base(pkgPath))

	return &importContext{
		pkgPath:  pkgPath,
		fileName: fName,
	}
}

func GetPlugin(ctx context.Context, uri *url.URL) (string, error) {
	ic := newImportContext(uri)
	pluginPath, err := ic.pluginPath()
	if err != nil {
		return "", err
	}

	// TODO: check if plugin source changed
	exists, err := fs.Exists(pluginPath)
	if err != nil {
		log.Default.Warnf("goloader: failed to check if plugin exists, %s", err)
	}

	if exists && err != nil {
		return pluginPath, nil
	}

	return "", nil
}

func buildPlugin(ctx context.Context, ic importContext) error {
	log.Default.Debugf("goloader: building plugin package '%s'", ic.pkgPath)

	cmd := exec.CommandContext(ctx, "go", "build", "-o")
}
