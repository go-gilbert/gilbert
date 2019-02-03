package tester

import (
	"errors"
	"os"
	"os/exec"
)

type Params struct {
	Items []TestEntry
}

type CoverageParams struct {
	Threshold       float32
	IgnoreUncovered bool
	Ignore          []string
}

type TestEntry struct {
	Path     string
	Coverage CoverageParams
}

func (e *TestEntry) ShouldCheckCoverage() bool {
	return e.Coverage.Threshold > 0
}

func (e *TestEntry) getTestingCommand(rootDir string) (cmd *exec.Cmd, tempFile *os.File, err error) {
	return nil, nil, errors.New("not implemented")
	//args := []string{"test", e.Path}
	//
	//if e.ShouldCheckCoverage() {
	//	tempFilePath := tempCoverFile(rootDir, e.Path)
	//	tempFile, err := os.OpenFile(tempFilePath, os.O_RDWR|os.O_CREATE, 0755)
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//
	//	args = append(args, "-coverprofile="+tempFilePath)
	//}
	//
	//cmd = exec.Command("go", args...)
	//return cmd, tempFile, err
}
