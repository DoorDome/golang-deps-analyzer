package modutils

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

type Parser interface {
	Parse(path string) (*Module, error)
}

type ModFileParser struct {
	repoRoot string
}

func NewModFileParser(repoRoot string) *ModFileParser {
	return &ModFileParser{repoRoot: repoRoot}
}

func (p *ModFileParser) Parse(path string) (*Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	f, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	dir := filepath.Dir(path)
	relDir, err := filepath.Rel(p.repoRoot, dir)
	if err != nil {
		relDir = dir
	}

	mod := &Module{
		Name:         f.Module.Mod.Path,
		AbsolutePath: dir,
		RelativePath: relDir,
	}
	if f.Go != nil {
		mod.GoVersion = f.Go.Version
	}
	for _, req := range f.Require {
		mod.Dependencies = append(mod.Dependencies, Dependency{
			Path:     req.Mod.Path,
			Version:  req.Mod.Version,
			Indirect: req.Indirect,
		})
	}
	return mod, nil
}
