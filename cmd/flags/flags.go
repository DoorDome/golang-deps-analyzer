package flags

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/pflag"
)

var ErrHelp = pflag.ErrHelp

type Config struct {
	RepoURL    string
	DirectOnly bool
	Timeout    time.Duration
	Format     string
	ProxyURL   string
}

func Parse(binaryName string, args []string, usageOut io.Writer) (*Config, error) {
	fs := pflag.NewFlagSet("deps-analyzer", pflag.ContinueOnError)
	fs.SetOutput(usageOut)
	fs.Usage = func() {
		fmt.Fprintf(usageOut, "Usage: %s [flags] <repository-url>\n\nFlags:\n", binaryName)
		fs.PrintDefaults()
	}

	cfg := &Config{}
	fs.BoolVar(&cfg.DirectOnly, "direct-only", false, "skip indirect dependencies when checking for updates")
	fs.DurationVar(&cfg.Timeout, "timeout", 5*time.Minute, "overall operation timeout")
	fs.StringVarP(&cfg.Format, "format", "f", "text", "output format: text or json")
	fs.StringVar(&cfg.ProxyURL, "proxy", "https://proxy.golang.org", "Go module proxy URL")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	remaining := fs.Args()
	if len(remaining) != 1 {
		fs.Usage()
		return nil, fmt.Errorf("expected exactly one repository URL, got %d argument(s)", len(remaining))
	}
	cfg.RepoURL = remaining[0]
	return cfg, nil
}
