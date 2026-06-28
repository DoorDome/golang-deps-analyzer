package modutils_test

import (
	"path/filepath"
	"testing"

	"deps-analyzer/internal/modutils"
)

const fullmodule = `module github.com/example/repo

go 1.21

require (
	github.com/foo/bar v1.2.3
	github.com/baz/qux v0.5.0 // indirect
)
`

func TestModFileParser_Parse(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), []byte(fullmodule))

	mod, err := modutils.NewModFileParser(root).Parse(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got, want := mod.Name, "github.com/example/repo"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := mod.GoVersion, "1.21"; got != want {
		t.Errorf("GoVersion = %q, want %q", got, want)
	}
	if got, want := mod.RelativePath, "."; got != want {
		t.Errorf("RelativePath = %q, want %q", got, want)
	}
	if got, want := len(mod.Dependencies), 2; got != want {
		t.Fatalf("len(Dependencies) = %d, want %d", got, want)
	}

	direct := mod.Dependencies[0]
	if direct.Path != "github.com/foo/bar" || direct.Version != "v1.2.3" || direct.Indirect {
		t.Errorf("Dependencies[0] = %+v, want {github.com/foo/bar v1.2.3 false}", direct)
	}
	indirect := mod.Dependencies[1]
	if indirect.Path != "github.com/baz/qux" || indirect.Version != "v0.5.0" || !indirect.Indirect {
		t.Errorf("Dependencies[1] = %+v, want {github.com/baz/qux v0.5.0 true}", indirect)
	}
}

func TestModFileParser_Parse_SubModule(t *testing.T) {
	root := t.TempDir()
	subPath := filepath.Join(root, "pkg", "sub", "go.mod")
	writeFile(t, subPath, []byte("module example.com/sub\n\ngo 1.21\n"))

	mod, err := modutils.NewModFileParser(root).Parse(subPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := mod.RelativePath, filepath.Join("pkg", "sub"); got != want {
		t.Errorf("RelDir = %q, want %q", got, want)
	}
}

func TestModFileParser_Parse_NoGoDirective(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), []byte("module example.com/test\n"))

	mod, err := modutils.NewModFileParser(root).Parse(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if mod.GoVersion != "" {
		t.Errorf("GoVersion = %q, want empty", mod.GoVersion)
	}
}

func TestModFileParser_Parse_NoDependencies(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"))

	mod, err := modutils.NewModFileParser(root).Parse(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(mod.Dependencies) != 0 {
		t.Errorf("Dependencies = %v, want empty", mod.Dependencies)
	}
}

func TestModFileParser_Parse_InvalidContent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), []byte("this is not valid go.mod content @@@@"))

	_, err := modutils.NewModFileParser(root).Parse(filepath.Join(root, "go.mod"))
	if err == nil {
		t.Error("Parse() expected error for invalid go.mod, got nil")
	}
}

func TestModFileParser_Parse_NonexistentFile(t *testing.T) {
	_, err := modutils.NewModFileParser("/tmp").Parse("/nonexistent/path/go.mod")
	if err == nil {
		t.Error("Parse() expected error for nonexistent file, got nil")
	}
}
