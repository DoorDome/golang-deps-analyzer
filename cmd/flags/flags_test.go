package flags_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"deps-analyzer/cmd/flags"
)

const binaryName = "deps-analyzer"

func TestParse_ValidArgs(t *testing.T) {
	cfg, err := flags.Parse(binaryName, []string{"https://github.com/example/repo"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.RepoURL != "https://github.com/example/repo" {
		t.Errorf("RepoURL = %q, want %q", cfg.RepoURL, "https://github.com/example/repo")
	}
	if cfg.DirectOnly {
		t.Error("DirectOnly should default to false")
	}
	if cfg.Format != "text" {
		t.Errorf("Format = %q, want %q", cfg.Format, "text")
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 5*time.Minute)
	}
}

func TestParse_AllFlags(t *testing.T) {
	cfg, err := flags.Parse(binaryName, []string{
		"--direct-only",
		"--timeout", "2m",
		"--format", "json",
		"https://github.com/example/repo",
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !cfg.DirectOnly {
		t.Error("DirectOnly should be true")
	}
	if cfg.Timeout != 2*time.Minute {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 2*time.Minute)
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q, want %q", cfg.Format, "json")
	}
}

func TestParse_MissingURL(t *testing.T) {
	_, err := flags.Parse(binaryName, []string{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Parse() expected error for missing URL, got nil")
	}
}

func TestParse_TooManyArgs(t *testing.T) {
	_, err := flags.Parse(binaryName, []string{"url1", "url2"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Parse() expected error for too many args, got nil")
	}
}

func TestParse_Help(t *testing.T) {
	var buf bytes.Buffer
	_, err := flags.Parse(binaryName, []string{"--help"}, &buf)
	if !errors.Is(err, flags.ErrHelp) {
		t.Errorf("Parse(--help) error = %v, want ErrHelp", err)
	}
	if buf.Len() == 0 {
		t.Error("expected usage text written to output, got nothing")
	}
}
