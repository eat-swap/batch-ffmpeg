package main

import (
	"io/fs"
	"path/filepath"
)

const (
	SrcDir = "."
)

func main() {
	filepath.Walk(SrcDir, func(path string, info fs.FileInfo, err error) error {
		return err
	})
}
