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

// Lookup check if file exists in current directory or in parent directories.
// Allows to locate file not in current but in parent directories.
//
// Accepts file name, start location and search deep count
func Lookup(name, startLocation string, deepCount int) (string, bool, error) {
	location := startLocation
	for i := 0; i < deepCount; i++ {
		fpath := filepath.Join(location, name)
		exists, err := Exists(fpath)
		if exists {
			return fpath, true, nil
		}

		if err != nil {
			return "", false, err
		}

		// abort, if we reached root directory
		parent := filepath.Dir(location)
		if parent == location {
			break
		}

		location = parent
	}

	return "", false, nil
}
