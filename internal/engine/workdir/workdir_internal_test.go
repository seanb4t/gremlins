package workdir

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyToSkipsPathThatVanishedBeforeStat(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	cd := NewCachedDealer(t.TempDir(), srcDir)

	gone := filepath.Join(srcDir, ".git", "index.lock")
	statErr := &fs.PathError{Op: "lstat", Path: gone, Err: fs.ErrNotExist}

	if err := cd.copyTo(dstDir)(gone, nil, statErr); err != nil {
		t.Fatalf("expected a vanished path to be skipped, got %v", err)
	}
}

func TestCopyToKeepsOtherWalkErrors(t *testing.T) {
	cd := NewCachedDealer(t.TempDir(), t.TempDir())

	walkErr := &fs.PathError{Op: "lstat", Path: "x", Err: fs.ErrPermission}

	if err := cd.copyTo(t.TempDir())("x", nil, walkErr); err == nil {
		t.Fatal("expected a permission error to fail the copy")
	}
}

func TestCopyPathSkipsFileThatVanishedAfterStat(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	src := filepath.Join(srcDir, "index.lock")
	if err := os.WriteFile(src, []byte("lock"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(src); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dstDir, "index.lock")
	if err := copyPath(src, dst, info); err != nil {
		t.Fatalf("expected a vanished file to be skipped, got %v", err)
	}
	if _, err := os.Lstat(dst); !os.IsNotExist(err) {
		t.Fatalf("expected no copy of a vanished file, got %v", err)
	}
}
