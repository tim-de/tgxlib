package tgxlib

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"
)

func cStrLen(buf []byte) int {
	for ix := 0; ix < len(buf); ix += 1 {
		if buf[ix] == 0 {
			return ix
		}
	}
	return len(buf)
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

func getSliceSegment[T interface{}](slice []T, segment_size, index uint) []T {
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

func CaseInsensitiveQuery(path string) string {
	query := ""
	for _, char := range path {
		if unicode.IsLetter(char) {
			query += fmt.Sprintf("[%c%c]", unicode.ToLower(char), unicode.ToUpper(char))
		} else {
			query += string(char)
		}
	}
	return query
}

func FindExistingPath(path string) (string, error) {
	pathparts := filepath.SplitList(path)
	existpath := ""
	for ix, part := range pathparts {
		testpath := filepath.Join(existpath, CaseInsensitiveQuery(part))
		matches, err := filepath.Glob(testpath)
		if err != nil {
			return "", err
		}
		if len(matches) > 1 {
			return "", fmt.Errorf("Too many results for %s", testpath)
		}
		if len(matches) == 1 {
			existpath = filepath.Join(existpath, matches[0])
			pathparts[ix] = filepath.Base(matches[0])
		} else {
			break
		}
	}
	return filepath.Join(pathparts ...), nil
}
