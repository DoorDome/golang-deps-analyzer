package git

import (
	"context"
	"fmt"
	"os/exec"
)

type Cloner interface {
	Clone(ctx context.Context, repoURL, targetDir string) error
}

type ExecCloner struct{}

func NewExecCloner() *ExecCloner {
	return &ExecCloner{}
}

func (e *ExecCloner) Clone(ctx context.Context, repoURL, targetDir string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", "--quiet", repoURL, targetDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone: %w\n%s", err, out)
	}
	return nil
}
