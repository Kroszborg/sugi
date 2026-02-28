package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// defaultTimeout is applied to all local git commands (status, log, diff, …).
	// Prevents hangs on corrupt repos or unexpected stdin prompts.
	defaultTimeout = 30 * time.Second

	// networkTimeout is applied to operations that talk to a remote
	// (push, pull, fetch).  Network latency can be high on slow connections.
	networkTimeout = 2 * time.Minute
)

// Client wraps git command execution for a specific repository.
type Client struct {
	RepoPath string
}

// NewClient creates a Client for the repository containing startPath.
// It walks up the directory tree looking for a .git directory.
func NewClient(startPath string) (*Client, error) {
	root, err := findGitRoot(startPath)
	if err != nil {
		return nil, err
	}
	return &Client{RepoPath: root}, nil
}

// findGitRoot walks upward from path looking for a .git directory.
func findGitRoot(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	for {
		info, err := os.Stat(filepath.Join(abs, ".git"))
		if err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("not a git repository (or any parent directories)")
		}
		abs = parent
	}
}

// runCtx executes a git command with the given context and returns trimmed stdout.
func (c *Client) runCtx(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.RepoPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}

// run executes a git command with the default 30 s timeout.
func (c *Client) run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return c.runCtx(ctx, args...)
}

// runSlow executes a git command with a 2-minute network timeout.
// Use for push, pull, fetch and other operations that contact a remote.
func (c *Client) runSlow(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), networkTimeout)
	defer cancel()
	return c.runCtx(ctx, args...)
}

// runLines executes a git command and splits stdout on newlines.
func (c *Client) runLines(args ...string) ([]string, error) {
	out, err := c.run(args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// RepoName returns the base directory name of the repository.
func (c *Client) RepoName() string {
	return filepath.Base(c.RepoPath)
}

// CurrentBranch returns the name of the currently checked-out branch.
func (c *Client) CurrentBranch() (string, error) {
	return c.run("rev-parse", "--abbrev-ref", "HEAD")
}

// Push pushes the current branch to its upstream remote.
func (c *Client) Push() error {
	_, err := c.runSlow("push")
	return err
}

// PushSetUpstream pushes and sets upstream for new branches.
func (c *Client) PushSetUpstream(remote, branch string) error {
	_, err := c.runSlow("push", "--set-upstream", remote, branch)
	return err
}

// PushForceWithLease pushes with --force-with-lease (safe force push for rebased branches).
func (c *Client) PushForceWithLease(remote, branch string) error {
	_, err := c.runSlow("push", "--force-with-lease", remote, branch)
	return err
}

// RunPublic exposes run for use in packages that wrap Client. Use sparingly.
func (c *Client) RunPublic(args ...string) (string, error) {
	return c.run(args...)
}

// Pull fetches and merges the upstream branch.
func (c *Client) Pull() error {
	_, err := c.runSlow("pull")
	return err
}

// Fetch fetches from all remotes.
func (c *Client) Fetch() error {
	_, err := c.runSlow("fetch", "--all", "--prune")
	return err
}

// buildCmd creates an exec.Command for the repo but does not run it.
// The caller owns the lifetime of the process and must manage cancellation.
func (c *Client) buildCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.RepoPath
	return cmd
}
