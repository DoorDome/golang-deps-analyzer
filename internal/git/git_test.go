package git_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"deps-analyzer/internal/git"
)

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestExecClonerClone_LocalRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	src := t.TempDir()
	mustGit(t, src, "init", "-b", "main")
	mustGit(t, src, "config", "user.email", "test@test.com")
	mustGit(t, src, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(src, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, src, "add", ".")
	mustGit(t, src, "commit", "-m", "init")

	dst := t.TempDir()
	if err := git.NewExecCloner().Clone(context.Background(), src, dst); err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "go.mod")); err != nil {
		t.Errorf("go.mod not found in cloned repo: %v", err)
	}
}

func TestExecClonerClone_InvalidURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	err := git.NewExecCloner().Clone(context.Background(), "/nonexistent/path/@@@@", t.TempDir())
	if err == nil {
		t.Error("expected error cloning invalid URL, got nil")
	}
}

func TestExecClonerClone_ContextCancelled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := git.NewExecCloner().Clone(ctx, "https://github.com/example/repo", t.TempDir())
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}
