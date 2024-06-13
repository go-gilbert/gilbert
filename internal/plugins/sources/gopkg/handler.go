package gopkg

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/plugins/support"
	"github.com/go-gilbert/gilbert/internal/storage"
	"github.com/go-gilbert/gilbert/internal/support/fs"
	"github.com/go-gilbert/gilbert/internal/support/shell"
)

type importContext struct {
	pkgPath  string
	fileName string
	filePath string
	rebuild  bool
}

const (
	// ProviderName is handler protocol name
	ProviderName = "go"

	rebuildParam = "rebuild"
)

func newImportContext(uri *url.URL) (*importContext, error) {
	pkgPath := filepath.Clean(filepath.Join(uri.Host + "/" + uri.Path))
	fName := support.AddPluginExtension(filepath.Base(pkgPath))

	hasher := md5.New()
	if _, err := hasher.Write([]byte(pkgPath)); err != nil {
		return nil, err
	}
	pluginDir, err := storage.LocalPath(storage.Plugins, hex.EncodeToString(hasher.Sum(nil)))
	if err != nil {
		return nil, err
	}

	rebuild, _ := strconv.ParseBool(uri.Query().Get(rebuildParam))
	return &importContext{
		pkgPath:  pkgPath,
		fileName: fName,
		filePath: filepath.Join(pluginDir, fName),
		rebuild:  rebuild,
	}, nil
}

// GetPlugin returns plugin from URL
func GetPlugin(ctx context.Context, uri *url.URL) (string, error) {
	ic, err := newImportContext(uri)
	if err != nil {
		return "", err
	}

	// TODO: automatically check if plugin source changed
	if pluginCached(ic) && !ic.rebuild {
		return ic.filePath, nil
	}

	if err := buildPlugin(ctx, ic); err != nil {
		return "", fmt.Errorf("failed to build plugin package (%s)", err)
	}

	return ic.filePath, nil
}

func buildPlugin(ctx context.Context, ic *importContext) error {
	log.Default.Debugf("goloader: building plugin package '%s'", ic.pkgPath)

	cmd := exec.CommandContext(ctx, "go", "build", "-buildmode", support.BuildMode, "-o", ic.filePath, ".")
	cmd.Dir = ic.pkgPath

	cmd.Stdout = log.Default
	cmd.Stderr = log.Default.ErrorWriter()
	log.Default.Debugf("goloader: exec '%s'", strings.Join(cmd.Args, " "))

	if err := runGoCommand(cmd); err != nil {
		return err
	}

	log.Default.Debug("goloader: build successful")
	return nil
}

////////////////////////////////////////
// Functions represented as variables //
// to make them mock-able.            //
////////////////////////////////////////

var pluginCached = func(ic *importContext) bool {
	exists, err := fs.Exists(ic.filePath)
	if err != nil {
		log.Default.Warnf("goloader: failed to check if plugin exists, %s", err)
	}

	if exists && err == nil {
		return true
	}

	return false
}

var runGoCommand = func(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	return nil
}
