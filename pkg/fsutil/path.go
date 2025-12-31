package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolvePath takes a path x and a base path, returning an absolute path.
// If x is absolute, it returns x as-is.
// If x is relative, it joins x with basePath and converts to absolute.
func ResolvePath(basePath, relPath string) string {
	// Check if x is already absolute
	if filepath.IsAbs(relPath) {
		return relPath
	}

	// x is relative, so join it with basePath
	joined := filepath.Join(basePath, relPath)

	// Convert to absolute path
	absPath, err := filepath.Abs(joined)
	if err != nil {
		return joined
	}

	return absPath
}

// DirExists checks if the given path exists and is a directory.
func DirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		// Path doesn't exist or other error occurred
		return err
	}

	if info.IsDir() {
		return nil
	}

	return fmt.Errorf("%q is not a directory", path)
}
