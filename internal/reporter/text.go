package reporter

import (
	"fmt"
	"io"

	"deps-analyzer/internal/update"
)

type TextReporter struct{}

func NewTextReporter() *TextReporter { return &TextReporter{} }

func (t *TextReporter) Report(w io.Writer, repoURL string, results []*update.Result) error {
	fmt.Fprintf(w, "Repository: %s\n\n", repoURL)

	totalUpdates := 0
	for _, r := range results {
		mod := r.Module
		loc := mod.RelativePath
		if loc == "" || loc == "." {
			loc = "root"
		}

		directCount := 0
		for _, dep := range mod.Dependencies {
			if !dep.Indirect {
				directCount++
			}
		}
		indirectCount := len(mod.Dependencies) - directCount

		fmt.Fprintf(w, "Module: %s (%s)\n", mod.Name, loc)
		if mod.GoVersion != "" {
			fmt.Fprintf(w, "  Go version:   %s\n", mod.GoVersion)
		}
		fmt.Fprintf(w, "  Dependencies: %d (%d direct, %d indirect)\n",
			len(mod.Dependencies), directCount, indirectCount)

		if len(r.Updates) == 0 {
			fmt.Fprintln(w, "  Updates:      none available")
		} else {
			fmt.Fprintf(w, "  Updates:      %d available\n", len(r.Updates))
			for _, u := range r.Updates {
				suffix := ""
				if u.Indirect {
					suffix = " (indirect)"
				}
				fmt.Fprintf(w, "    %-55s %s -> %s%s\n", u.Path, u.Current, u.Latest, suffix)
			}
		}
		totalUpdates += len(r.Updates)
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "Summary: %d module(s) scanned, %d update(s) available\n",
		len(results), totalUpdates)
	return nil
}
