package tgxlib

import (
	"io/fs"
	"time"
)

type subFileInfo struct {
	name string
	fileLen int64
	fileMode fs.FileMode
	modTime time.Time
}

func (info subFileInfo) Name() string {
	return info.name
}

func (info subFileInfo) Size() int64 {
	return info.fileLen
}

func (info subFileInfo) Mode() fs.FileMode {
	var ret fs.FileMode
	return ret
}

func (info subFileInfo) IsDir() bool {
	return info.Mode().IsDir()
}

func (info subFileInfo) Sys() any {
	return nil
}
