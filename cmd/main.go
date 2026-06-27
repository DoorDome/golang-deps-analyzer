package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"deps-analyzer/cmd/flags"
	"deps-analyzer/internal/git"
	"deps-analyzer/internal/modutils"
	"deps-analyzer/internal/reporter"
	"deps-analyzer/internal/update"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := flags.Parse(os.Args[0], os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flags.ErrHelp) {
			return nil
		}
		return err
	}

	var rep reporter.Reporter
	switch cfg.Format {
	case "json":
		rep = reporter.NewJSONReporter()
	default:
		rep = reporter.NewTextReporter()
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	tmpDir, err := os.MkdirTemp("", "deps-analyzer-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Fprintf(os.Stderr, "Cloning %s...\n", cfg.RepoURL)
	if err := git.NewExecCloner().Clone(ctx, cfg.RepoURL, tmpDir); err != nil {
		return fmt.Errorf("clone repository: %w", err)
	}

	modPaths, err := modutils.NewFSFinder().Find(ctx, tmpDir)
	if err != nil {
		return fmt.Errorf("find modules: %w", err)
	}
	if len(modPaths) == 0 {
		fmt.Fprintln(os.Stderr, "No Go modules found in repository")
		return nil
	}
	fmt.Fprintf(os.Stderr, "Found %d module(s)\n", len(modPaths))

	parser := modutils.NewModFileParser(tmpDir)
	var modules []*modutils.Module
	for _, path := range modPaths {
		mod, err := parser.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			continue
		}
		modules = append(modules, mod)
	}

	fmt.Fprintln(os.Stderr, "Checking for updates...")
	updater := update.NewProxyUpdater().
		WithHTTPClient(&http.Client{Timeout: 30 * time.Second}).
		WithDirectOnly(cfg.DirectOnly).
		Done()

	var results []*update.Result
	for _, mod := range modules {
		fmt.Fprintf(os.Stderr, "  Checking %s...\n", mod.Name)
		result, err := updater.FindUpdates(ctx, mod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: find updates %s: %v\n", mod.Name, err)
			continue
		}

		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  warning: %v\n", e)
		}
		results = append(results, result)
	}

	return rep.Report(os.Stdout, cfg.RepoURL, results)
}
