package reporter

import (
	"encoding/json"
	"io"

	"deps-analyzer/internal/update"
)

type JSONReporter struct{}

func NewJSONReporter() *JSONReporter { return &JSONReporter{} }

type jsonOutput struct {
	Repository     string       `json:"repository"`
	Modules        []jsonModule `json:"modules"`
	ModulesScanned int          `json:"modules_scanned"`
	TotalUpdates   int          `json:"total_updates"`
}

type jsonModule struct {
	Path                 string       `json:"path"`
	Location             string       `json:"location"`
	GoVersion            string       `json:"go_version,omitempty"`
	TotalDependencies    int          `json:"total_dependencies"`
	DirectDependencies   int          `json:"direct_dependencies"`
	IndirectDependencies int          `json:"indirect_dependencies"`
	Updates              []jsonUpdate `json:"updates"`
}

type jsonUpdate struct {
	Path     string `json:"path"`
	Current  string `json:"current"`
	Latest   string `json:"latest"`
	Indirect bool   `json:"indirect"`
}

func (j *JSONReporter) Report(w io.Writer, repoURL string, results []*update.Result) error {
	out := jsonOutput{
		Repository: repoURL,
		Modules:    make([]jsonModule, 0, len(results)),
	}

	for _, r := range results {
		mod := r.Module
		directCount := 0
		for _, dep := range mod.Dependencies {
			if !dep.Indirect {
				directCount++
			}
		}

		updates := make([]jsonUpdate, 0, len(r.Updates))
		for _, u := range r.Updates {
			updates = append(updates, jsonUpdate{
				Path:     u.Path,
				Current:  u.Current,
				Latest:   u.Latest,
				Indirect: u.Indirect,
			})
		}

		out.Modules = append(out.Modules, jsonModule{
			Path:                 mod.Name,
			Location:             mod.RelativePath,
			GoVersion:            mod.GoVersion,
			TotalDependencies:    len(mod.Dependencies),
			DirectDependencies:   directCount,
			IndirectDependencies: len(mod.Dependencies) - directCount,
			Updates:              updates,
		})
		out.TotalUpdates += len(r.Updates)
	}
	out.ModulesScanned = len(results)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
