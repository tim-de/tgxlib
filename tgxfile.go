package tgxlib

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type TgxFile struct {
	Version VersionNumber
	Identifier string
	Checksum uint32
	FileLen uint32
	FileCount uint32
	SubFiles []subfile
}

func (file TgxFile) String() string {
	return fmt.Sprintf("Version: %v\nChecksum: 0x%08x\nLength: 0x%08x\nFileCount: %d\nSubfiles: %+v", file.Version, file.Checksum, file.FileLen, file.FileCount, file.SubFiles)
}

func findNextFileStart(offset uint32) uint32 {
	return (offset & 0xfffff800) + 0x800
}

func (file TgxFile) computeFileLength() uint32 {
	var offset uint32 = (132 * file.FileCount) + 116
	for _, sub := range file.SubFiles {
		offset = findNextFileStart(offset)
		offset += sub.fileLen
	}
	return offset 
}

func (file *TgxFile) setFileLength() {
	file.FileLen = file.computeFileLength()
}

func (file *TgxFile) ComputeLengthAndOffsets() {
	var offset uint32 = (132 * file.FileCount) + 116
	var header_offset uint32 = 0
	for ix, subfile := range file.SubFiles {
		offset = findNextFileStart(offset)
		if subfile.headerLen != 0 {
			file.SubFiles[ix].headerOffset = header_offset
			header_offset += subfile.headerLen
		}
		file.SubFiles[ix].fileIndex = uint32(ix)
		file.SubFiles[ix].startOffset = offset
		offset += subfile.fileLen
		file.SubFiles[ix].endOffset = offset
	}
	file.FileLen = offset
}

func (file TgxFile) CreateHeader() []byte {
	header_length := (132 * file.FileCount) + 116
	output_buffer := make([]byte, header_length)

	file.ComputeLengthAndOffsets()

	binary.LittleEndian.PutUint32(output_buffer[0x0:], 0x0001000f)
	binary.LittleEndian.PutUint32(output_buffer[0x8:], 0xfa7e843f)
	binary.LittleEndian.PutUint32(output_buffer[0xc:], file.Version.pack())
	binary.LittleEndian.PutUint32(output_buffer[0x14:], file.FileLen)

	constant_section := [12]byte{0xAA, 0x97, 0xA7, 0xEF, 0x52, 0x13, 0x83, 0xE1, 0xD7, 0xCD, 0xC1, 0x85}
	copy(output_buffer[0x18:], constant_section[:])

	copy(output_buffer[0x24:], []byte(file.Identifier))
	
	filespec_offset := uint32(0x74)
	filelen_offset := filespec_offset + (file.FileCount * 104)
	filepos_offset := filelen_offset + (file.FileCount * 20)

	binary.LittleEndian.PutUint32(output_buffer[0x3c:], filespec_offset)
	binary.LittleEndian.PutUint32(output_buffer[0x40:], file.FileCount)
	binary.LittleEndian.PutUint32(output_buffer[0x44:], filelen_offset)
	binary.LittleEndian.PutUint32(output_buffer[0x48:], file.FileCount)
	binary.LittleEndian.PutUint32(output_buffer[0x4c:], filepos_offset)
	binary.LittleEndian.PutUint32(output_buffer[0x50:], file.FileCount)

	filespec_buf := output_buffer[filespec_offset:filelen_offset]
	filelen_buf := output_buffer[filelen_offset:filepos_offset]
	filepos_buf := output_buffer[filepos_offset:]

	for ix, subfile := range file.SubFiles {
		subfile.fileIndex = uint32(ix)
		subfile.writeFileSpec(getSliceSegment[byte](filespec_buf, 104, uint(subfile.fileIndex)))
		subfile.writeLenSpec(getSliceSegment[byte](filelen_buf, 20, uint(subfile.fileIndex)))
		subfile.writePosSpec(getSliceSegment[byte](filepos_buf, 8, uint(subfile.fileIndex)))
	}

	return output_buffer
}

func ReadFromFile(file_path string) (TgxFile, error) {
	extension := strings.ToUpper(filepath.Ext(file_path))
	result := TgxFile{}
	if extension != ".TGX" && extension != ".TGW" {
		return result, fmt.Errorf("%v is not a .tgx or .tgw file", file_path)
	}
	file_handle, err := os.Open(file_path)
	if err != nil {
		return result, err
	}
	defer file_handle.Close()
	header_buffer := make([]byte, 116)
	_, err = file_handle.Read(header_buffer)
	if err != nil {
		return result, err
	}
	packedversion := binary.LittleEndian.Uint32(header_buffer[0xc:])
	result.Version = unpackVersionFromU32(packedversion)

	result.Checksum = binary.LittleEndian.Uint32(header_buffer[0x10:])

	result.FileLen = binary.LittleEndian.Uint32(header_buffer[0x14:])

	result.Identifier = string(header_buffer[0x24:0x28])

	filespec_offset := binary.LittleEndian.Uint32(header_buffer[0x3c:])
	filespec_count := binary.LittleEndian.Uint32(header_buffer[0x40:])
	filelength_offset := binary.LittleEndian.Uint32(header_buffer[0x44:])
	filelength_count := binary.LittleEndian.Uint32(header_buffer[0x48:])
	filepos_offset := binary.LittleEndian.Uint32(header_buffer[0x4c:])
	filepos_count := binary.LittleEndian.Uint32(header_buffer[0x50:])

	result.FileCount = filespec_count

	filespec_buflen := 104 * filespec_count
	filelen_buflen := 20 * filelength_count
	filepos_buflen := 8 * filepos_count

	filespec_buffer := make([]byte, filespec_buflen)
	filelen_buffer := make([]byte, filelen_buflen)
	filepos_buffer := make([]byte, filepos_buflen)

	_, err = file_handle.Seek(int64(filespec_offset), 0)
	if err != nil {
		return TgxFile{}, err
	}
	_, err = file_handle.Read(filespec_buffer)
	if err != nil {
		return TgxFile{}, err
	}

	_, err = file_handle.Seek(int64(filelength_offset), 0)
	if err != nil {
		return TgxFile{}, err
	}
	_, err = file_handle.Read(filelen_buffer)
	if err != nil {
		return TgxFile{}, err
	}

	_, err = file_handle.Seek(int64(filepos_offset), 0)
	if err != nil {
		return TgxFile{}, err
	}
	_, err = file_handle.Read(filepos_buffer)
	if err != nil {
		return TgxFile{}, err
	}

	result.SubFiles = make([]subfile, result.FileCount)

	var file_ix uint
	for file_ix = 0; file_ix < uint(result.FileCount); file_ix += 1 {
		raw_filespec := getSliceSegment[byte](filespec_buffer, 104, file_ix)
		filespec_ix := binary.LittleEndian.Uint32(raw_filespec[92:])
		raw_filelength := getSliceSegment[byte](filelen_buffer, 20, file_ix)
		filelen_ix := binary.LittleEndian.Uint32(raw_filelength[16:])
		raw_filepos := getSliceSegment[byte](filepos_buffer, 8, uint(filelen_ix))
		result.SubFiles[filespec_ix].readFileSpec(raw_filespec)
		result.SubFiles[filelen_ix].readFilePos(raw_filepos)
		_, err = file_handle.Seek(int64(result.SubFiles[filelen_ix].startOffset), 0)
		if err != nil {
			return TgxFile{}, err
		}
		_, err = file_handle.Read(result.SubFiles[filelen_ix].fileData)
		if err != nil {
			return TgxFile{}, err
		}
	}
	return result, nil
}

func FromPathList(version VersionNumber, short_id string, pathlist []string) (TgxFile, error) {
	var result TgxFile
	var err error

	result.Version = version
	result.Identifier = short_id
	result.FileCount = uint32(len(pathlist))
	result.SubFiles = make([]subfile, result.FileCount)

	for ix, path := range pathlist {
		//fmt.Fprintln(os.Stderr, path)
		result.SubFiles[ix], err = ImportSubfile(path)
		if err != nil {
			return TgxFile{}, err
		}
	}
	result.ComputeLengthAndOffsets()

	return result, nil
}

func (file TgxFile) WriteFile(outfile_path string) error {

	file_handle, err := os.Create(outfile_path)
	if err != nil {
		return err
	}
	defer file_handle.Close()
	
	header := file.CreateHeader()
	_, err = file_handle.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = file_handle.Write(header)
	if err != nil {
		return err
	}

	for _, subfile := range file.SubFiles {
		/*subfilepath, err := filepath.Rel(".", subfile.FilePath)
		if err != nil {
			return err
		}
		filebuf, err := os.ReadFile(subfilepath)
		if err != nil {
			return err
		}*/
		_, err = file_handle.Seek(int64(subfile.startOffset), 0)
		if err != nil {
			return err
		}
		_, err = file_handle.Write(subfile.fileData)
		if err != nil {
			return err
		}		
	}
	return nil
}
