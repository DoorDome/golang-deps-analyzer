package reporter

import (
	"io"

	"deps-analyzer/internal/update"
)

type Reporter interface {
	Report(w io.Writer, repoURL string, results []*update.Result) error
}
