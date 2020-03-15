package main

import (
	"fmt"
	"os"

	"github.com/go-gilbert/gilbert/v2/manifest"
)

const fname = "gilbert.hcl"

func main() {
	man, err := manifest.FromFile(fname, nil)
	if err != nil {
		switch t := err.(type) {
		case *manifest.Error:
			fmt.Println(t.PrettyPrint())
		default:
			fmt.Println(err)
		}
		os.Exit(1)
	}

	fmt.Println(man)
}

func must(err error) {
	if err == nil {
		return
	}

	panic(err)
}
