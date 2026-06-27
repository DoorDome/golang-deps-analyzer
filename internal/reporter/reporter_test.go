package reporter_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"deps-analyzer/internal/modutils"
	"deps-analyzer/internal/reporter"
	"deps-analyzer/internal/update"
)

var testResults = []*update.Result{
	{
		Module: &modutils.Module{
			Name:         "github.com/example/root",
			GoVersion:    "1.21",
			RelativePath: ".",
			Dependencies: []modutils.Dependency{
				{Path: "github.com/foo/bar", Version: "v1.0.0"},
				{Path: "github.com/baz/qux", Version: "v0.5.0", Indirect: true},
			},
		},
		Updates: []update.Update{
			{Path: "github.com/foo/bar", Current: "v1.0.0", Latest: "v1.5.0"},
		},
	},
	{
		Module: &modutils.Module{
			Name:         "github.com/example/sub",
			RelativePath: "pkg/sub",
			Dependencies: []modutils.Dependency{},
		},
		Updates: []update.Update{},
	},
}

func TestTextReporter_WithUpdates(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.NewTextReporter()
	if err := rep.Report(&buf, "https://github.com/example/repo", testResults); err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"Repository: https://github.com/example/repo",
		"Module: github.com/example/root (root)",
		"Go version:   1.21",
		"Dependencies: 2 (1 direct, 1 indirect)",
		"Updates:      1 available",
		"github.com/foo/bar",
		"v1.0.0 -> v1.5.0",
		"Module: github.com/example/sub (pkg/sub)",
		"Updates:      none available",
		"Summary: 2 module(s) scanned, 1 update(s) available",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestTextReporter_ImplementsInterface(t *testing.T) {
	var _ reporter.Reporter = reporter.NewTextReporter()
}

func TestJSONReporter_ImplementsInterface(t *testing.T) {
	var _ reporter.Reporter = reporter.NewJSONReporter()
}

func TestJSONReporter_Output(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.NewJSONReporter()
	if err := rep.Report(&buf, "https://github.com/example/repo", testResults); err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	var out struct {
		Repository     string `json:"repository"`
		ModulesScanned int    `json:"modules_scanned"`
		TotalUpdates   int    `json:"total_updates"`
		Modules        []struct {
			Path                 string `json:"path"`
			Location             string `json:"location"`
			GoVersion            string `json:"go_version"`
			TotalDependencies    int    `json:"total_dependencies"`
			DirectDependencies   int    `json:"direct_dependencies"`
			IndirectDependencies int    `json:"indirect_dependencies"`
			Updates              []struct {
				Path     string `json:"path"`
				Current  string `json:"current"`
				Latest   string `json:"latest"`
				Indirect bool   `json:"indirect"`
			} `json:"updates"`
		} `json:"modules"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw:\n%s", err, buf.String())
	}

	if out.Repository != "https://github.com/example/repo" {
		t.Errorf("repository = %q", out.Repository)
	}
	if out.ModulesScanned != 2 {
		t.Errorf("modules_scanned = %d, want 2", out.ModulesScanned)
	}
	if out.TotalUpdates != 1 {
		t.Errorf("total_updates = %d, want 1", out.TotalUpdates)
	}
	if len(out.Modules) != 2 {
		t.Fatalf("len(modules) = %d, want 2", len(out.Modules))
	}

	root := out.Modules[0]
	if root.Path != "github.com/example/root" {
		t.Errorf("modules[0].path = %q", root.Path)
	}
	if root.TotalDependencies != 2 || root.DirectDependencies != 1 || root.IndirectDependencies != 1 {
		t.Errorf("dependency counts: total=%d direct=%d indirect=%d",
			root.TotalDependencies, root.DirectDependencies, root.IndirectDependencies)
	}
	if len(root.Updates) != 1 {
		t.Fatalf("modules[0].updates len = %d, want 1", len(root.Updates))
	}
	u := root.Updates[0]
	if u.Path != "github.com/foo/bar" || u.Current != "v1.0.0" || u.Latest != "v1.5.0" {
		t.Errorf("update = %+v", u)
	}
}
