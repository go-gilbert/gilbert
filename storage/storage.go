package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-gilbert/gilbert/support/fs"
)

const (
	homeDirName = ".gilbert"

	// StoreVarName is env var name for storage path
	StoreVarName = "GILBERT_HOME"
)

var storageDir string

// Type represents storage type
type Type int

const (
	// Root is storage root
	Root = iota

	// Plugins represents plugins storage
	Plugins
)

var storageTypes = map[Type]string{
	Root:    "",
	Plugins: "plugins",
}

func home() (string, error) {
	if storageDir != "" {
		return storageDir, nil
	}

	// override storage directory by env variable if present
	if envVal := os.Getenv(StoreVarName); envVal != "" {
		storageDir = envVal
		return storageDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get storage directory, %s", err)
	}

	return filepath.Join(home, homeDirName), nil
}

// Path returns storage item path
func Path(storageType Type, paths ...string) (string, error) {
	home, err := home()
	if err != nil {
		return "", err
	}

	dir, ok := storageTypes[storageType]
	if !ok {
		return "", errors.New("unknown storage type")
	}

	p := filepath.Join(home, dir)

	if len(paths) > 0 {
		p += "/" + filepath.Join(paths...)
	}

	return p, nil
}

// LocalPath returns local project storage path by type
func LocalPath(storageType Type, paths ...string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir, ok := storageTypes[storageType]
	if !ok {
		return "", errors.New("unknown storage type")
	}

	p := filepath.Join(wd, homeDirName, dir)

	if len(paths) > 0 {
		p += "/" + filepath.Join(paths...)
	}

	return p, nil
}

// Delete clears specified storage item
func Delete(storageType Type, paths ...string) error {
	dir, err := Path(storageType, paths...)
	if err != nil {
		return err
	}

	exists, err := fs.Exists(dir)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	return os.RemoveAll(dir)
}
