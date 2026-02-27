package git

import (
	"os"
	"path/filepath"
	"strings"
)

// BisectStatus holds the current state of a git bisect session.
type BisectStatus struct {
	InProgress  bool
	CurrentHash string // the commit bisect is currently testing
	GoodCount   int
	BadCount    int
	Log         []string // recent bisect log lines
}

// BisectInProgress returns true if a bisect session is active.
func (c *Client) BisectInProgress() bool {
	_, err := os.Stat(filepath.Join(c.RepoPath, ".git", "BISECT_HEAD"))
	if err == nil {
		return true
	}
	_, err = os.Stat(filepath.Join(c.RepoPath, ".git", "BISECT_START"))
	return err == nil
}

// BisectStart begins a bisect session.
func (c *Client) BisectStart() error {
	_, err := c.run("bisect", "start")
	return err
}

// BisectGood marks a commit as good. If hash is "", marks HEAD good.
func (c *Client) BisectGood(hash string) error {
	if hash == "" {
		_, err := c.run("bisect", "good")
		return err
	}
	_, err := c.run("bisect", "good", hash)
	return err
}

// BisectBad marks a commit as bad. If hash is "", marks HEAD bad.
func (c *Client) BisectBad(hash string) error {
	if hash == "" {
		_, err := c.run("bisect", "bad")
		return err
	}
	_, err := c.run("bisect", "bad", hash)
	return err
}

// BisectSkip skips the current commit.
func (c *Client) BisectSkip() error {
	_, err := c.run("bisect", "skip")
	return err
}

// BisectReset ends the bisect session and returns to the original branch.
func (c *Client) BisectReset() error {
	_, err := c.run("bisect", "reset")
	return err
}

// BisectLog returns the bisect log lines.
func (c *Client) BisectLog() ([]string, error) {
	out, err := c.run("bisect", "log")
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimRight(l, "\r")
		if l != "" && !strings.HasPrefix(l, "#") {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

// BisectGetStatus returns a snapshot of the current bisect state.
func (c *Client) BisectGetStatus() BisectStatus {
	if !c.BisectInProgress() {
		return BisectStatus{}
	}

	status := BisectStatus{InProgress: true}

	// Current commit being tested
	hash, _ := c.run("rev-parse", "--short", "BISECT_HEAD")
	status.CurrentHash = strings.TrimSpace(hash)

	// Parse log for good/bad counts
	logLines, _ := c.BisectLog()
	status.Log = logLines
	for _, l := range logLines {
		if strings.HasPrefix(l, "git bisect good") {
			status.GoodCount++
		} else if strings.HasPrefix(l, "git bisect bad") {
			status.BadCount++
		}
	}

	return status
}
