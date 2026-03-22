// Package fsutil provides small helpers for creating directories and writing files
// used when building the local website mirror.
package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Common os.OpenFile flag combinations for mirror output.
const (
	FsCWAFlags = os.O_CREATE | os.O_WRONLY | os.O_APPEND // create if needed, append, write-only
	FsCWTFlags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC  // create if needed, truncate, write-only
	FsCWFlags  = os.O_CREATE | os.O_WRONLY               // create if needed, write-only (no truncate)
	FsRFlags   = os.O_RDONLY                             // read-only
)

// OpenFile opens a file like [os.OpenFile] and ensures the parent directory exists
// (via [os.MkdirAll]) before opening.
func OpenFile(filePath string, flag int, perm os.FileMode) (*os.File, error) {
	fileDir := filepath.Dir(filePath)
	if err := os.MkdirAll(fileDir, 0o775); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filePath, flag, perm)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// WriteOSFile writes data to f and closes f. data must be []byte, string, or [io.Reader];
// any other type returns an error after closing f.
func WriteOSFile(f *os.File, data any) (n int, err error) {
	switch typData := data.(type) {
	case []byte:
		n, err = f.Write(typData)
	case string:
		n, err = f.WriteString(typData)
	case io.Reader:
		var n64 int64
		n64, err = io.Copy(f, typData)
		n = int(n64)
	default:
		_ = f.Close()
		return 0, fmt.Errorf("WriteOSFile: unsupported data type %T", data)
	}

	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return n, err
}
