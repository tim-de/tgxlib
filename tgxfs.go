package tgxlib

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type tgxFS struct {
	version VersionNumber
	twoCharacterID string
	subfiles []fs.DirEntry
	data []byte
}

func (ds *tgxFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == "." {
		return ds.subfiles, nil
	}
	
}

func LoadAsFS(sourcepath string) (tgxFS, error) {
	extension := strings.ToUpper(filepath.Ext(sourcepath))
	if extension != ".TGX" && extension != ".TGW" {
		return tgxFS{}, fmt.Errorf("%s is not a .tgx or .tgw file", sourcepath)
	}
	
	var fs tgxFS
	var err error
	fs.data, err = os.ReadFile(sourcepath)
	if err != nil {
		return tgxFS{}, err
	}

	// TODO:
	// Parse header section to create the file entries
	packedversion := binary.LittleEndian.Uint32(fs.data[0xc:])
	fs.version = unpackVersionFromU32(packedversion)

	fs.twoCharacterID = string(fs.data[0x24:0x28])

	filespec_offset := binary.LittleEndian.Uint32(fs.data[0x3c:])
	filespec_count := binary.LittleEndian.Uint32(fs.data[0x40:])
	filelength_offset := binary.LittleEndian.Uint32(fs.data[0x44:])
	//filelength_count := binary.LittleEndian.Uint32(fs.data[0x48:])
	filepos_offset := binary.LittleEndian.Uint32(fs.data[0x4c:])
	filepos_count := binary.LittleEndian.Uint32(fs.data[0x50:])

	//filespec_buflen := 104 * filespec_count
	//filelen_buflen := 20 * filelength_count
	filepos_buflen := 8 * filepos_count

	filespec_buffer := fs.data[filespec_offset:filelength_offset]
	filelen_buffer := fs.data[filelength_offset:filepos_offset]
	filepos_buffer := fs.data[filepos_offset:filepos_offset+filepos_buflen]

	// Now read each file spec section and create the different
	// file/dir entries required

	for file_ix := 0; file_ix < int(filespec_count); file_ix ++ {
		this_filespec := getSliceSegment(filespec_buffer, 104, file_ix)
		this_filelength := getSliceSegment(filelen_buffer, 20, file_ix)

		filelength_ix := binary.LittleEndian.Uint32(this_filelength[16:])
		this_filepos := getSliceSegment(filepos_buffer, 8, int(filelength_ix))
		fs.addFileEntry(this_filespec, this_filepos)
	}
	
	return fs, nil
}

func binarySearchIn(name string, filelist []fs.DirEntry) (fs.DirEntry, error) {
	if len(filelist) == 0 {
		return nil, fs.ErrNotExist
	}
	pivot := len(filelist) / 2
	switch strings.Compare(name, filelist[pivot].Name()) {
	case 0:
		return filelist[pivot], nil
	case 1:
		return binarySearchIn(name, filelist[:pivot])
	case -1:
		return binarySearchIn(name, filelist[pivot:])
	}
	return nil, fs.ErrInvalid
}

func findFileIx(name string, filelist []fs.DirEntry) (int, error) {
	sublist := filelist
	var ix int
	for pivot := len(sublist)/2; pivot > 0; pivot /= 2 {
		switch strings.Compare(name, sublist[pivot].Name()) {
		case 0:
			return ix + pivot, nil
		case 1:
			sublist = sublist[:pivot]
		case 2:
			ix += pivot
			sublist = sublist[pivot:]
		}
	}
}

func (ds *tgxFS) findFile(name string) (fs.DirEntry, error) {
	cwd := "."
	filelist, err := ds.ReadDir(cwd)
	if err != nil {
		return nil, err
	}
	path_elements := strings.Split(name, string(filepath.Separator))
	var entry fs.DirEntry
	for ix, element := range path_elements {
		entry, err = binarySearchIn(element, filelist)
		if err != nil {
			return nil, err
		}
		filelist, err = ds.ReadDir(filepath.Join(path_elements[:ix]...))
		if err != nil {
			return nil, err
		}
	}
	return entry, nil
}

func (ds *tgxFS) MkDir(name string) error {
	newdirname := filepath.Base(name)
	parentdirname := filepath.Dir(name)
	parentdir, err := ds.findFile(parentdirname)
}

func (ds tgxFS) addFileEntry(filespec, filepos []byte) {
	var newEntry tgxFileEntry
	path_len := cStrLen(filespec)
	file_path := InternalToOsPath(string(filespec[:path_len]))
	newEntry.name = filepath.Base(file_path)
	newEntry.start = binary.LittleEndian.Uint32(filepos)
	newEntry.end = binary.LittleEndian.Uint32(filepos[4:])
}

type tgxDirEntry struct {
	name string
	subfiles []fs.DirEntry
}
	
func (dir tgxDirEntry) ReadDir(n int) ([]fs.DirEntry, error) {
	if len(dir.subfiles) == 0 {
		return nil, io.EOF
	}
	if n <= 0 {
		n = len(dir.subfiles)
	}
	return dir.subfiles[:n], nil
}

//    Structure of filesystem load from a tgx file:
// 1. Read the entire file into a slice
// 2. Grok the header, and walk each filepath by element,
//    adding a new directory for every branch node and a
//    new file for every leaf node
