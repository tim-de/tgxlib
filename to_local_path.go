//go:build !windows

package tgxlib

import (
	"strings"
)

func InternalToOsPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func OsToInternalPath(path string) string {
	return strings.ReplaceAll(path, "/", "\\")
}
