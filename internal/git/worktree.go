package git

import (
	"strings"
)

// Worktree represents a git worktree.
type Worktree struct {
	Path      string
	Branch    string
	Head      string
	IsBare    bool
	IsMain    bool
	IsLocked  bool
}

// Worktrees returns all worktrees for this repository.
func (c *Client) Worktrees() ([]Worktree, error) {
	out, err := c.run("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktrees(out), nil
}

// AddWorktree creates a new worktree at path for the given branch.
func (c *Client) AddWorktree(path, branch string, newBranch bool) error {
	args := []string{"worktree", "add"}
	if newBranch {
		args = append(args, "-b", branch)
	}
	args = append(args, path)
	if !newBranch {
		args = append(args, branch)
	}
	_, err := c.run(args...)
	return err
}

// RemoveWorktree removes the worktree at the given path.
func (c *Client) RemoveWorktree(path string) error {
	_, err := c.run("worktree", "remove", path)
	return err
}

// ForceRemoveWorktree forcibly removes a worktree (even if dirty).
func (c *Client) ForceRemoveWorktree(path string) error {
	_, err := c.run("worktree", "remove", "--force", path)
	return err
}

// parseWorktrees parses git worktree list --porcelain output.
func parseWorktrees(raw string) []Worktree {
	var worktrees []Worktree
	var current *Worktree

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			if current != nil {
				worktrees = append(worktrees, *current)
				current = nil
			}
			continue
		}

		if current == nil {
			current = &Worktree{}
			worktrees = append(worktrees, *current)
			current = &worktrees[len(worktrees)-1]
		}

		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			b := strings.TrimPrefix(line, "branch ")
			// Strip "refs/heads/" prefix
			current.Branch = strings.TrimPrefix(b, "refs/heads/")
		case line == "bare":
			current.IsBare = true
		case line == "locked":
			current.IsLocked = true
		}
	}

	if current != nil && current.Path != "" {
		// Check if already appended
		found := false
		for _, w := range worktrees {
			if w.Path == current.Path {
				found = true
				break
			}
		}
		if !found {
			worktrees = append(worktrees, *current)
		}
	}

	// Mark main worktree (first one)
	if len(worktrees) > 0 {
		worktrees[0].IsMain = true
	}

	return worktrees
}
