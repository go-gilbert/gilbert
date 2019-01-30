package tools

import (
	"path/filepath"
	"strings"
)

func TrimFileExtension(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}
