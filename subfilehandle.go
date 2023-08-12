package tgxlib

import (
	"io/fs"
	"path/filepath"
	"time"
)

type subfileHandle struct {
	header *subfile
	fileData []byte
	readOffset int
}

// Implement fs.File interface
func (file *subfileHandle) Stat() (fs.FileInfo, error) {
	info := subFileInfo{
		name: filepath.Base(file.header.FilePath),
		fileLen: int64(len(file.fileData)),
		modTime: time.Time{},
	}
	return info, nil
}

func (file *subfileHandle) Read(buf []byte) (int, error)  {
	numBytes := copy(buf, file.fileData[file.readOffset:])
	file.readOffset += numBytes
	return numBytes, nil
}

func (file *subfileHandle) Close() error {
	return nil
}
// end of fs.File implementation
