package git

import (
	"strings"
)

// Stage adds the given file paths to the index.
func (c *Client) Stage(paths ...string) error {
	args := append([]string{"add", "--"}, paths...)
	_, err := c.run(args...)
	return err
}

// StageAll stages all changes including untracked files.
func (c *Client) StageAll() error {
	_, err := c.run("add", "-A")
	return err
}

// Unstage removes the given file paths from the index (keeps working tree).
func (c *Client) Unstage(paths ...string) error {
	args := append([]string{"restore", "--staged", "--"}, paths...)
	_, err := c.run(args...)
	return err
}

// UnstageAll removes all staged changes from the index.
func (c *Client) UnstageAll() error {
	_, err := c.run("restore", "--staged", ".")
	return err
}

// Commit creates a commit with the given message.
func (c *Client) Commit(message string) error {
	_, err := c.run("commit", "-m", message)
	return err
}

// CommitAmend amends the last commit with the current index and optional new message.
func (c *Client) CommitAmend(message string) error {
	if message == "" {
		_, err := c.run("commit", "--amend", "--no-edit")
		return err
	}
	_, err := c.run("commit", "--amend", "-m", message)
	return err
}

// DiscardFile discards all unstaged changes in a file (restores from HEAD).
func (c *Client) DiscardFile(path string) error {
	_, err := c.run("restore", "--", path)
	return err
}

// StashPush creates a new stash entry with an optional message.
func (c *Client) StashPush(message string) error {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	_, err := c.run(args...)
	return err
}

// StashPop applies and drops the top stash entry.
func (c *Client) StashPop() error {
	_, err := c.run("stash", "pop")
	return err
}

// StashList returns all stash entries.
func (c *Client) StashList() ([]string, error) {
	return c.runLines("stash", "list")
}

// StageHunk applies a specific diff hunk to the index via stdin.
// patch should be a valid unified diff patch string.
func (c *Client) StageHunk(patch string) error {
	cmd := c.buildCmd("apply", "--cached", "--unidiff-zero", "-")
	cmd.Stdin = strings.NewReader(patch)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// ResetSoft moves HEAD to ref, keeping staged and working-tree changes.
func (c *Client) ResetSoft(ref string) error {
	_, err := c.run("reset", "--soft", ref)
	return err
}

// ResetMixed moves HEAD to ref, unstaging changes but keeping working tree.
func (c *Client) ResetMixed(ref string) error {
	_, err := c.run("reset", "--mixed", ref)
	return err
}

// ResetHard moves HEAD to ref, discarding all staged and unstaged changes.
func (c *Client) ResetHard(ref string) error {
	_, err := c.run("reset", "--hard", ref)
	return err
}

// UnstageHunk reverses a specific diff hunk from the index.
func (c *Client) UnstageHunk(patch string) error {
	cmd := c.buildCmd("apply", "--cached", "--unidiff-zero", "--reverse", "-")
	cmd.Stdin = strings.NewReader(patch)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
