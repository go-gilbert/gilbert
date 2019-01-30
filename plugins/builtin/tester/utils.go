package tester

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"path"
)

func tempCoverFile(dir, pkgPath string) string {
	h := md5.New()
	io.WriteString(h, pkgPath)
	return path.Join(dir, string(h.Sum(nil))+".cover.out")
}

func tempDir() (string, error) {
	dir, err := ioutil.TempDir("guru-testrunner", "")
	if err != nil {
		err = fmt.Errorf("failed to create temp dir: %v", err)
	}

	return dir, err
}
