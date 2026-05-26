package fs

import (
	"os"
	"path/filepath"
	"strings"
)

// TrimFileExtension removes extension from file name
func TrimFileExtension(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}

// Exists check if current file exists
func Exists(location string) (bool, error) {
	_, err := os.Stat(location)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
