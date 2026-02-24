package git

import (
	"strings"
)

// Branch represents a git branch.
type Branch struct {
	Name      string
	IsCurrent bool
	IsRemote  bool
	Upstream  string
	Ahead     int
	Behind    int
}

// Branches returns all local branches.
func (c *Client) Branches() ([]Branch, error) {
	// --format with tracking info
	lines, err := c.runLines(
		"branch", "--list",
		"--format=%(refname:short)\t%(upstream:short)\t%(ahead-behind:HEAD)",
		"--sort=-committerdate",
	)
	if err != nil {
		return nil, err
	}

	current, _ := c.CurrentBranch()

	var branches []Branch
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		b := Branch{Name: parts[0]}
		b.IsCurrent = b.Name == current

		if len(parts) > 1 {
			b.Upstream = parts[1]
		}
		// ahead-behind from HEAD is not what we want here; skip parsing it
		// as it compares to HEAD not upstream. We'll use the simpler version.
		branches = append(branches, b)
	}
	return branches, nil
}

// BranchesWithTracking returns branches with ahead/behind counts vs upstream.
func (c *Client) BranchesWithTracking() ([]Branch, error) {
	branches, err := c.Branches()
	if err != nil {
		return nil, err
	}

	for i, b := range branches {
		if b.Upstream == "" {
			continue
		}
		out, err := c.run("rev-list", "--left-right", "--count",
			b.Name+"..."+b.Upstream)
		if err != nil {
			continue
		}
		parts := strings.Fields(out)
		if len(parts) == 2 {
			ahead, behind := 0, 0
			for _, r := range parts[0] {
				ahead = ahead*10 + int(r-'0')
			}
			for _, r := range parts[1] {
				behind = behind*10 + int(r-'0')
			}
			branches[i].Ahead = ahead
			branches[i].Behind = behind
		}
	}
	return branches, nil
}

// Checkout switches to the named branch.
func (c *Client) Checkout(name string) error {
	_, err := c.run("checkout", name)
	return err
}

// CreateBranch creates and switches to a new branch from HEAD.
func (c *Client) CreateBranch(name string) error {
	_, err := c.run("checkout", "-b", name)
	return err
}

// DeleteBranch deletes a local branch (must be fully merged).
func (c *Client) DeleteBranch(name string) error {
	_, err := c.run("branch", "-d", name)
	return err
}

// ForceDeleteBranch deletes a local branch regardless of merge status.
func (c *Client) ForceDeleteBranch(name string) error {
	_, err := c.run("branch", "-D", name)
	return err
}

// RenameBranch renames the current branch.
func (c *Client) RenameBranch(newName string) error {
	_, err := c.run("branch", "-m", newName)
	return err
}
