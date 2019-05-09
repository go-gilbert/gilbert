package gopkg

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/plugins/support"
	"github.com/go-gilbert/gilbert/storage"
	"github.com/go-gilbert/gilbert/support/fs"
	"github.com/go-gilbert/gilbert/support/shell"
)

type importContext struct {
	pkgPath  string
	fileName string
	filePath string
}

// ProviderName is handler protocol name
const ProviderName = "go"

func newImportContext(uri *url.URL) (*importContext, error) {
	pkgPath := filepath.Clean(filepath.Join(uri.Host + "/" + uri.Path))
	fName := support.AddPluginExtension(filepath.Base(pkgPath))

	hasher := md5.New()
	hasher.Write([]byte(pkgPath))
	pluginDir, err := storage.LocalPath(storage.Plugins, hex.EncodeToString(hasher.Sum(nil)))
	if err != nil {
		return nil, err
	}

	return &importContext{
		pkgPath:  pkgPath,
		fileName: fName,
		filePath: filepath.Join(pluginDir, fName),
	}, nil
}

func GetPlugin(ctx context.Context, uri *url.URL) (string, error) {
	ic, err := newImportContext(uri)
	if err != nil {
		return "", err
	}

	// TODO: check if plugin source changed
	exists, err := fs.Exists(ic.filePath)
	if err != nil {
		log.Default.Warnf("goloader: failed to check if plugin exists, %s", err)
	}

	if exists && err != nil {
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

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	log.Default.Debug("goloader: build successful")
	return nil
}
