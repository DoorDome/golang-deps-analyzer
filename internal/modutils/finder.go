package modutils

import (
	"context"
	"io/fs"
	"path/filepath"
)

type Finder interface {
	Find(ctx context.Context, root string) ([]string, error)
}

type FSFinder struct{}

func NewFSFinder() *FSFinder {
	return &FSFinder{}
}

// Find returns the absolute paths of all go.mod files under root. Hidden directories (starting
// with '.') are skipped.
func (f *FSFinder) Find(ctx context.Context, root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			switch name := d.Name(); {
			case len(name) > 0 && name[0] == '.':
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "go.mod" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
