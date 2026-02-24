package git

import (
	"strings"
	"time"
)

// StashEntry represents a single stash.
type StashEntry struct {
	Index   int
	Ref     string // e.g. "stash@{0}"
	Branch  string // branch stash was created from
	Message string
	Date    time.Time
}

// Stashes returns all stash entries.
func (c *Client) Stashes() ([]StashEntry, error) {
	sep := "\x1f"
	format := strings.Join([]string{"%gd", "%gs", "%ci"}, sep)
	lines, err := c.runLines("stash", "list", "--format="+format)
	if err != nil {
		return nil, err
	}

	var entries []StashEntry
	for i, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, sep, 3)
		if len(parts) < 2 {
			continue
		}

		ref := parts[0]
		msg := parts[1]
		date := time.Time{}
		if len(parts) > 2 {
			date, _ = time.Parse("2006-01-02 15:04:05 -0700", parts[2])
		}

		// Extract branch from message "On branch: message" or "WIP on branch: message"
		branch := ""
		msgClean := msg
		if strings.HasPrefix(msg, "WIP on ") {
			rest := strings.TrimPrefix(msg, "WIP on ")
			if idx := strings.Index(rest, ":"); idx >= 0 {
				branch = rest[:idx]
				msgClean = strings.TrimSpace(rest[idx+1:])
			}
		} else if strings.HasPrefix(msg, "On ") {
			rest := strings.TrimPrefix(msg, "On ")
			if idx := strings.Index(rest, ":"); idx >= 0 {
				branch = rest[:idx]
				msgClean = strings.TrimSpace(rest[idx+1:])
			}
		}

		entries = append(entries, StashEntry{
			Index:   i,
			Ref:     ref,
			Branch:  branch,
			Message: msgClean,
			Date:    date,
		})
	}
	return entries, nil
}

// StashShow returns the diff for a stash entry.
func (c *Client) StashShow(ref string) ([]FileDiff, error) {
	out, err := c.run("stash", "show", "-p", "--unified=3", ref)
	if err != nil {
		return nil, err
	}
	return parseDiff(out), nil
}

// StashDrop drops a stash entry by ref.
func (c *Client) StashDrop(ref string) error {
	_, err := c.run("stash", "drop", ref)
	return err
}

// StashApply applies a stash entry without dropping it.
func (c *Client) StashApply(ref string) error {
	_, err := c.run("stash", "apply", ref)
	return err
}
