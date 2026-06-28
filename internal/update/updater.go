package update

import (
	"context"

	"deps-analyzer/internal/modutils"
)

type Update struct {
	Path     string
	Current  string
	Latest   string
	Indirect bool
}

type Result struct {
	Module  *modutils.Module
	Updates []Update
	Errors  []error
}

type Updater interface {
	FindUpdates(ctx context.Context, mod *modutils.Module) (*Result, error)
}
