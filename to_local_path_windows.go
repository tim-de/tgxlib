//go:build windows

package tgxlib

func InternalToOsPath(path string) string {
	return path
}

func OsToInternalPath(path string) string {
	return path
}
