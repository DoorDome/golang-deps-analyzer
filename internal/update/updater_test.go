package update_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"deps-analyzer/internal/modutils"
	"deps-analyzer/internal/update"
)

func makeProxyServer(versions map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		path = strings.TrimSuffix(path, "/@latest")
		ver, ok := versions[path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]string{"Version": ver}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
}

func newUpdaterBuilder(srv *httptest.Server) *update.ProxyUpdaterBuilder {
	return update.NewProxyUpdater().
		WithProxyURL(srv.URL).
		WithHTTPClient(srv.Client())
}

func TestProxyChecker_Check_HasUpdates(t *testing.T) {
	srv := makeProxyServer(map[string]string{
		"github.com/foo/bar": "v1.5.0",
		"github.com/baz/qux": "v0.5.0",
	})
	defer srv.Close()

	mod := &modutils.Module{
		Name: "github.com/example/repo",
		Dependencies: []modutils.Dependency{
			{Path: "github.com/foo/bar", Version: "v1.2.0"},
			{Path: "github.com/baz/qux", Version: "v0.5.0"},
		},
	}

	result, err := newUpdaterBuilder(srv).Done().FindUpdates(context.Background(), mod)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if got, want := len(result.Updates), 1; got != want {
		t.Fatalf("len(Updates) = %d, want %d: %+v", got, want, result.Updates)
	}
	u := result.Updates[0]
	if u.Path != "github.com/foo/bar" || u.Current != "v1.2.0" || u.Latest != "v1.5.0" {
		t.Errorf("unexpected update: %+v", u)
	}
}

func TestProxyChecker_Check_NoUpdates(t *testing.T) {
	srv := makeProxyServer(map[string]string{
		"github.com/foo/bar": "v1.0.0",
	})
	defer srv.Close()

	mod := &modutils.Module{
		Dependencies: []modutils.Dependency{
			{Path: "github.com/foo/bar", Version: "v1.0.0"},
		},
	}

	result, err := newUpdaterBuilder(srv).Done().FindUpdates(context.Background(), mod)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(result.Updates) != 0 {
		t.Errorf("expected no updates, got %+v", result.Updates)
	}
}

func TestProxyChecker_Check_DirectOnly(t *testing.T) {
	srv := makeProxyServer(map[string]string{
		"github.com/foo/bar": "v2.0.0",
		"github.com/baz/qux": "v2.0.0",
	})
	defer srv.Close()

	mod := &modutils.Module{
		Dependencies: []modutils.Dependency{
			{Path: "github.com/foo/bar", Version: "v1.0.0", Indirect: false},
			{Path: "github.com/baz/qux", Version: "v1.0.0", Indirect: true},
		},
	}

	result, err := newUpdaterBuilder(srv).
		WithDirectOnly(true).
		Done().
		FindUpdates(context.Background(), mod)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if got, want := len(result.Updates), 1; got != want {
		t.Fatalf("len(Updates) = %d, want %d (direct only)", got, want)
	}
	if result.Updates[0].Path != "github.com/foo/bar" {
		t.Errorf("expected direct dep update, got %s", result.Updates[0].Path)
	}
}

func TestProxyChecker_Check_ProxyNotFound(t *testing.T) {
	srv := makeProxyServer(map[string]string{})
	defer srv.Close()

	mod := &modutils.Module{
		Dependencies: []modutils.Dependency{
			{Path: "github.com/foo/bar", Version: "v1.0.0"},
		},
	}

	result, err := newUpdaterBuilder(srv).Done().FindUpdates(context.Background(), mod)
	if err != nil {
		t.Fatalf("Check() must not return error for proxy 404: %v", err)
	}
	if len(result.Updates) != 0 {
		t.Errorf("expected no updates on 404, got %+v", result.Updates)
	}
	if len(result.Errors) == 0 {
		t.Error("expected Errors to be populated for proxy 404")
	}
}

func TestProxyChecker_Check_SkipsInvalidSemver(t *testing.T) {
	srv := makeProxyServer(map[string]string{})
	defer srv.Close()

	mod := &modutils.Module{
		Dependencies: []modutils.Dependency{
			{Path: "github.com/foo/bar", Version: "not-a-semver"},
		},
	}

	result, err := newUpdaterBuilder(srv).Done().FindUpdates(context.Background(), mod)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(result.Updates) != 0 || len(result.Errors) != 0 {
		t.Errorf("expected empty result for invalid semver, got updates=%v errors=%v",
			result.Updates, result.Errors)
	}
}

func TestProxyChecker_Check_SkipsPseudoVersion(t *testing.T) {
	// v1.1.4-0.20250930002214-15135a999495 is a commit after the v1.1.4 tag.
	// The proxy returns v1.1.4 as "latest", which semver considers greater than
	// the pseudo-version (pre-release < release). Reporting this as an update
	// would be a false positive so we skip pseudo-versions.
	srv := makeProxyServer(map[string]string{
		"github.com/foo/bar": "v1.1.4",
	})
	defer srv.Close()

	mod := &modutils.Module{
		Dependencies: []modutils.Dependency{
			{Path: "github.com/foo/bar", Version: "v1.1.4-0.20250930002214-15135a999495"},
		},
	}

	result, err := newUpdaterBuilder(srv).Done().FindUpdates(context.Background(), mod)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(result.Updates) != 0 {
		t.Errorf("pseudo-version should not be reported as updatable, got %+v", result.Updates)
	}
}

func TestProxyChecker_Check_EmptyDependencies(t *testing.T) {
	srv := makeProxyServer(map[string]string{})
	defer srv.Close()

	mod := &modutils.Module{Dependencies: []modutils.Dependency{}}

	result, err := newUpdaterBuilder(srv).Done().FindUpdates(context.Background(), mod)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(result.Updates) != 0 || len(result.Errors) != 0 {
		t.Errorf("expected empty result for no dependencies")
	}
}

func TestProxyChecker_Check_IndirectFlagged(t *testing.T) {
	srv := makeProxyServer(map[string]string{
		"github.com/foo/bar": "v2.0.0",
	})
	defer srv.Close()

	mod := &modutils.Module{
		Dependencies: []modutils.Dependency{
			{Path: "github.com/foo/bar", Version: "v1.0.0", Indirect: true},
		},
	}

	result, err := newUpdaterBuilder(srv).Done().FindUpdates(context.Background(), mod)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(result.Updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(result.Updates))
	}
	if !result.Updates[0].Indirect {
		t.Error("Update.Indirect should be true for indirect dependency")
	}
}
