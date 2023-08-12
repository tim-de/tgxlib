package tgxlib

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type subfile struct {
	FilePath string
	fileLen uint32
	fileIdentifier uint32
	fileIndex uint32
	headerOffset uint32
	headerLen uint32
	startOffset uint32
	endOffset uint32
	fileHandle fs.File
}

// Implement fs.DirEntry interface
func (file subfile) Name() string {
	return filepath.Base(file.FilePath)
}

// Implement sort interface
type subfileList []subfile
func (files subfileList) Len() int {
	return len(files)
}

func (files subfileList) Swap(i, j int) {
	files[i], files[j] = files[j], files[i]
}

func (files subfileList) Less(i, j int) bool {
	return files[i].fileIdentifier < files[j].fileIdentifier
}
// End of sort implementation

func (file *subfile) setIdentifier() {
	file.fileIdentifier = genIdentifier(OsToInternalPath(file.FilePath))
}

func (file subfile) getHeaderLength() uint32 {
	switch strings.ToUpper(filepath.Ext(file.FilePath)) {
	case ".WAV":
		return 0x24
	default:
		return 0x0
	}
}

func fromFileSpec(file_spec_buf []byte) subfile {
	var result subfile
	result.readFileSpec(file_spec_buf)
	return result
}

func ImportSubfile(file_path string) (subfile, error) {
	var result subfile
	result.FilePath = file_path
	result.setIdentifier()
	//fmt.Fprintf(os.Stderr, "%s => 0x%08x\n", result.FilePath, result.fileIdentifier)
	fileinfo, err := os.Stat(file_path)
	if err != nil {
		return subfile{}, err
	}
	result.fileLen = uint32(fileinfo.Size())
	result.headerLen = result.getHeaderLength()
	result.fileHandle, err = os.Open(file_path)
	if err != nil {
		return subfile{}, err
	}
	return result, nil
}

func (file *subfile) readFileSpec(file_spec_buf []byte) {
	path_len := cStrLen(file_spec_buf)
	file.FilePath = InternalToOsPath(string(file_spec_buf[:path_len]))
	file.fileIdentifier = binary.LittleEndian.Uint32(file_spec_buf[0x50:])
	file.fileLen = binary.LittleEndian.Uint32(file_spec_buf[0x54:])
	file.fileIndex = binary.LittleEndian.Uint32(file_spec_buf[0x5c:])
	file.headerOffset = binary.LittleEndian.Uint32(file_spec_buf[0x60:])
	file.headerLen = binary.LittleEndian.Uint32(file_spec_buf[0x64:])
}

func (file *subfile) readFilePos(file_pos_buf, tgxbuf []byte) {
	file.startOffset = binary.LittleEndian.Uint32(file_pos_buf)
	file.endOffset = binary.LittleEndian.Uint32(file_pos_buf[0x4:])
	file.fileHandle = &subfileHandle{
		header: file,
		fileData: tgxbuf[file.startOffset:file.endOffset],
		readOffset: 0,
	}
}

func (file subfile) writeFileSpec(buf []byte) error {
	if len(file.FilePath) >= 80 {
		return fmt.Errorf("File path %s is too long", file.FilePath)
	}

	copy(buf, []byte(OsToInternalPath(file.FilePath)))
	buf[len(file.FilePath)] = 0

	binary.LittleEndian.PutUint32(buf[80:], file.fileIdentifier)
	binary.LittleEndian.PutUint32(buf[84:], file.fileLen)
	binary.LittleEndian.PutUint32(buf[88:], 1)
	binary.LittleEndian.PutUint32(buf[92:], file.fileIndex)
	binary.LittleEndian.PutUint32(buf[96:], file.headerOffset)
	binary.LittleEndian.PutUint32(buf[100:], file.headerLen)
	return nil
}

func (file subfile) writeLenSpec(buf []byte) {
	binary.LittleEndian.PutUint32(buf[0x8:], file.fileLen)
	binary.LittleEndian.PutUint32(buf[0xc:], 1)
	binary.LittleEndian.PutUint32(buf[0x10:], file.fileIndex)
}

func (file subfile) writePosSpec(buf []byte) {
	binary.LittleEndian.PutUint32(buf[0x0:], file.startOffset)
	binary.LittleEndian.PutUint32(buf[0x4:], file.endOffset)
}

func (file subfile) Dump(rootdir string) error {
	rootdir = filepath.Clean(rootdir)
	outfilename := filepath.Join(rootdir, strings.ToUpper(file.FilePath))
	subdir := filepath.Dir(outfilename)
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		err := os.MkdirAll(subdir, 0700)
		if err != nil {
			return err
		}
	}

	outfilehandle, err := os.Create(outfilename)
	if err != nil {
		return err
	}
	defer outfilehandle.Close()

	data := make([]byte, file.fileLen)
	if _, err := file.fileHandle.Read(data); err != nil {
		return err
	}
	_, err = outfilehandle.Write(data)
	if err != nil {
		return err
	}
	return nil
}
