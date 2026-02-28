package git

import (
	"strings"
	"time"
)

// Tag represents a git tag.
type Tag struct {
	Name        string
	Hash        string
	ShortHash   string
	Date        time.Time
	Message     string // for annotated tags
	IsAnnotated bool
}

// Tags returns all local tags sorted by creation date descending.
func (c *Client) Tags() ([]Tag, error) {
	sep := "\x1f"
	format := "%(refname:short)" + sep +
		"%(objectname:short)" + sep +
		"%(creatordate:iso)" + sep +
		"%(subject)"

	lines, err := c.runLines(
		"tag", "--list",
		"--format="+format,
		"--sort=-creatordate",
	)
	if err != nil {
		return nil, err
	}

	var tags []Tag
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, sep)
		if len(parts) < 2 {
			continue
		}
		t := Tag{
			Name:      parts[0],
			ShortHash: parts[1],
		}
		if len(parts) > 2 {
			t.Date, _ = time.Parse("2006-01-02 15:04:05 -0700", parts[2])
		}
		if len(parts) > 3 {
			t.Message = strings.TrimSpace(parts[3])
			t.IsAnnotated = t.Message != ""
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// CreateTag creates a lightweight tag at HEAD.
func (c *Client) CreateTag(name string) error {
	_, err := c.run("tag", name)
	return err
}

// CreateAnnotatedTag creates an annotated tag at HEAD with a message.
func (c *Client) CreateAnnotatedTag(name, message string) error {
	_, err := c.run("tag", "-a", name, "-m", message)
	return err
}

// DeleteTag deletes a local tag.
func (c *Client) DeleteTag(name string) error {
	_, err := c.run("tag", "-d", name)
	return err
}

// PushTag pushes a tag to the remote.
func (c *Client) PushTag(remote, name string) error {
	_, err := c.run("push", remote, "refs/tags/"+name)
	return err
}

// DeleteRemoteTag deletes a tag from the remote.
func (c *Client) DeleteRemoteTag(remote, name string) error {
	_, err := c.run("push", remote, "--delete", "refs/tags/"+name)
	return err
}

// CherryPick applies the commit with the given hash to the current branch.
func (c *Client) CherryPick(hash string) error {
	_, err := c.run("cherry-pick", hash)
	return err
}

// CherryPickAbort aborts an in-progress cherry-pick.
func (c *Client) CherryPickAbort() error {
	_, err := c.run("cherry-pick", "--abort")
	return err
}

// RevertCommit creates a new commit that undoes the changes from hash.
func (c *Client) RevertCommit(hash string) error {
	_, err := c.run("revert", "--no-edit", hash)
	return err
}
