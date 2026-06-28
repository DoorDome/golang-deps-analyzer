package modutils_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"deps-analyzer/internal/modutils"
)

var minimalMod = []byte("module example.com/test\n\ngo 1.21\n")

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFSFinder_Find_Basic(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), minimalMod)
	writeFile(t, filepath.Join(root, "sub", "go.mod"), minimalMod)
	writeFile(t, filepath.Join(root, "sub", "nested", "go.mod"), minimalMod)

	found, err := modutils.NewFSFinder().Find(context.Background(), root)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if got, want := len(found), 3; got != want {
		t.Errorf("Find() = %d paths, want %d: %v", got, want, found)
	}
}

func TestFSFinder_Find_SkipsHiddenDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), minimalMod)
	writeFile(t, filepath.Join(root, ".git", "go.mod"), minimalMod)
	writeFile(t, filepath.Join(root, ".hidden", "go.mod"), minimalMod)

	found, err := modutils.NewFSFinder().Find(context.Background(), root)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if got, want := len(found), 1; got != want {
		t.Errorf("Find() = %d paths, want %d (hidden dirs must be skipped): %v", got, want, found)
	}
}

func TestFSFinder_Find_Empty(t *testing.T) {
	root := t.TempDir()
	found, err := modutils.NewFSFinder().Find(context.Background(), root)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(found) != 0 {
		t.Errorf("Find() on empty dir = %d paths, want 0", len(found))
	}
}

func TestFSFinder_Find_ContextCancelled(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), minimalMod)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := modutils.NewFSFinder().Find(ctx, root)
	if err == nil {
		t.Error("Find() with cancelled context should return error")
	}
}
