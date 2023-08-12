package tgxlib

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func cStrLen(buf []byte) int {
	for ix := 0; ix < len(buf); ix += 1 {
		if buf[ix] == 0 {
			return ix
		}
	}
	return len(buf)
}

func sliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for ix, value := range a {
		if b[ix] != value {
			return false
		}
	}
	return true
}

func insert[T any](slice []T, index int, element T) []T {
	slice = append(slice[:index+1], slice[index:]...)
	slice[index] = element
	return slice
}

func genIdentifier(filePath string) uint32 {
	if filePath == "" {
		return 0
	}
	filePath = strings.ToUpper(filePath)
	//var rune_list = []rune(strings.ToUpper(filePath))
	var result uint32 = uint32(filePath[0]) << 8

	for ix, char := range filePath[1:] {
		result += ((result >> 4) * uint32(char)) + uint32(ix)
	}
	return result
}

func getSliceSegment[T any](slice []T, segment_size, index int) []T {
	return slice[index * segment_size:(index + 1) * segment_size]
}

func FindFilesRecursive(root string) ([]string, error) {
	pathlist := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() == false {
			pathlist = append(pathlist, path)
		}
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return pathlist, nil
}
