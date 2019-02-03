package fs

import (
	"path/filepath"
	"strings"
)

// TrimFileExtension removes extension from file name
func TrimFileExtension(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}
