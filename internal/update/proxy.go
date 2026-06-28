package update

import (
	"context"
	"deps-analyzer/internal/modutils"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

type ProxyUpdater struct {
	httpClient *http.Client
	proxyURL   string
	directOnly bool
}

type ProxyUpdaterBuilder struct {
	httpClient *http.Client
	proxyURL   string
	directOnly bool
}

func NewProxyUpdater() *ProxyUpdaterBuilder {
	return &ProxyUpdaterBuilder{
		httpClient: &http.Client{},
		proxyURL:   "https://proxy.golang.org",
	}
}

func (b *ProxyUpdaterBuilder) WithHTTPClient(client *http.Client) *ProxyUpdaterBuilder {
	b.httpClient = client
	return b
}

func (b *ProxyUpdaterBuilder) WithProxyURL(url string) *ProxyUpdaterBuilder {
	b.proxyURL = url
	return b
}

func (b *ProxyUpdaterBuilder) WithDirectOnly(directOnly bool) *ProxyUpdaterBuilder {
	b.directOnly = directOnly
	return b
}

func (b *ProxyUpdaterBuilder) Done() *ProxyUpdater {
	return &ProxyUpdater{
		httpClient: b.httpClient,
		proxyURL:   b.proxyURL,
		directOnly: b.directOnly,
	}
}

type latestResponse struct {
	Version string `json:"Version"`
}

func (c *ProxyUpdater) latestVersion(ctx context.Context, modPath string) (string, error) {
	escaped, err := module.EscapePath(modPath)
	if err != nil {
		return "", fmt.Errorf("escape path %s: %w", modPath, err)
	}
	url := c.proxyURL + "/" + escaped + "/@latest"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound, http.StatusGone:
		return "", fmt.Errorf("module %s not available on proxy (HTTP %d)", modPath, resp.StatusCode)
	default:
		return "", fmt.Errorf("proxy returned HTTP %d for %s", resp.StatusCode, modPath)
	}

	var info latestResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("decode proxy response for %s: %w", modPath, err)
	}
	if info.Version == "" {
		return "", fmt.Errorf("proxy returned empty version for %s", modPath)
	}
	return info.Version, nil
}

func (c *ProxyUpdater) FindUpdates(ctx context.Context, mod *modutils.Module) (*Result, error) {
	result := &Result{Module: mod}
	for _, dep := range mod.Dependencies {
		if c.directOnly && dep.Indirect {
			continue
		}
		if !semver.IsValid(dep.Version) || module.IsPseudoVersion(dep.Version) {
			continue
		}

		latest, err := c.latestVersion(ctx, dep.Path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", dep.Path, err))
			continue
		}
		if !semver.IsValid(latest) {
			continue
		}

		if semver.Compare(latest, dep.Version) > 0 {
			result.Updates = append(result.Updates, Update{
				Path:     dep.Path,
				Current:  dep.Version,
				Latest:   latest,
				Indirect: dep.Indirect,
			})
		}
	}
	return result, nil
}
