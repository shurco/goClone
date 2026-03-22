package fsutil

import (
	"os"
	"path/filepath"
)

// IsDir reports whether the named directory exists.
func IsDir(path string) bool {
	if path == "" || len(path) > 468 {
		return false
	}

	if fi, err := os.Stat(path); err == nil {
		return fi.IsDir()
	}
	return false
}

// Workdir returns the process current working directory, or "" if [os.Getwd] fails.
func Workdir() string {
	dir, _ := os.Getwd()
	return dir
}

// MkDirs ensures each path in dirPaths exists as a directory, creating missing paths
// with perm. Paths that already exist as directories are left unchanged.
func MkDirs(perm os.FileMode, dirPaths ...string) error {
	for _, dirPath := range dirPaths {
		if !IsDir(dirPath) {
			if err := os.MkdirAll(dirPath, perm); err != nil {
				return err
			}
		}
	}
	return nil
}

// MkSubDirs creates parentDir/name for each name in subDirs when that path is not
// already an existing directory.
func MkSubDirs(perm os.FileMode, parentDir string, subDirs ...string) error {
	for _, dirName := range subDirs {
		dirPath := filepath.Join(parentDir, dirName)
		if !IsDir(dirPath) {
			if err := os.MkdirAll(dirPath, perm); err != nil {
				return err
			}
		}
	}
	return nil
}
