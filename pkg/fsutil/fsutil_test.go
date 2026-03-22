package fsutil

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func Test_OpenFile_CreatesParents(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "a/b/c.txt")
	f, err := OpenFile(p, FsCWFlags, 0o666)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = f.Close()
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func Test_WriteOSFile_VariousTypes(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "x.txt")
	f, err := OpenFile(p, FsCWFlags, 0o666)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := WriteOSFile(f, "hello"); err != nil {
		t.Fatalf("write string: %v", err)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "hello" {
		t.Fatalf("unexpected content: %q", string(b))
	}

	p = filepath.Join(tmp, "y.txt")
	f, err = OpenFile(p, FsCWFlags, 0o666)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := WriteOSFile(f, []byte("world")); err != nil {
		t.Fatalf("write bytes: %v", err)
	}
	b, _ = os.ReadFile(p)
	if string(b) != "world" {
		t.Fatalf("unexpected content: %q", string(b))
	}

	p = filepath.Join(tmp, "z.txt")
	f, err = OpenFile(p, FsCWFlags, 0o666)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := WriteOSFile(f, bytes.NewBufferString("buf")); err != nil {
		t.Fatalf("write reader: %v", err)
	}
	b, _ = os.ReadFile(p)
	if string(b) != "buf" {
		t.Fatalf("unexpected content: %q", string(b))
	}
}

func Test_WriteOSFile_Truncates(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "trunc.txt")

	f, _ := OpenFile(p, FsCWTFlags, 0o666)
	_, _ = WriteOSFile(f, "longer content here")

	f2, err := OpenFile(p, FsCWTFlags, 0o666)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if _, err := WriteOSFile(f2, "short"); err != nil {
		t.Fatalf("write: %v", err)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "short" {
		t.Fatalf("expected truncation, got %q", string(b))
	}
}

func Test_WriteOSFile_UnknownTypeReturnsError(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "err.txt")
	f, err := OpenFile(p, FsCWFlags, 0o666)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_, err = WriteOSFile(f, 12345)
	if err == nil {
		t.Fatalf("expected error for unsupported type, got nil")
	}
}
