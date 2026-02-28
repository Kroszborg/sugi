package git

import (
	"strconv"
	"strings"
	"time"
)

// ReflogEntry represents a single reflog entry.
type ReflogEntry struct {
	Hash      string
	ShortHash string
	Ref       string // e.g. "HEAD@{0}"
	Action    string // e.g. "commit", "checkout", "rebase"
	Subject   string
	Date      time.Time
}

// Reflog returns the reflog for HEAD.
func (c *Client) Reflog(limit int) ([]ReflogEntry, error) {
	sep := "\x1f"
	format := strings.Join([]string{"%H", "%h", "%gd", "%gs", "%ci"}, sep)
	args := []string{"reflog", "--format=" + format}
	if limit > 0 {
		args = append(args, "--max-count", strconv.Itoa(limit))
	}

	lines, err := c.runLines(args...)
	if err != nil {
		return nil, err
	}

	var entries []ReflogEntry
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, sep, 5)
		if len(parts) < 4 {
			continue
		}

		date := time.Time{}
		if len(parts) > 4 {
			date, _ = time.Parse("2006-01-02 15:04:05 -0700", parts[4])
		}

		// Parse action from subject: "action: subject" or just "action"
		action, subject := parseReflogSubject(parts[3])

		entries = append(entries, ReflogEntry{
			Hash:      parts[0],
			ShortHash: parts[1],
			Ref:       parts[2],
			Action:    action,
			Subject:   subject,
			Date:      date,
		})
	}
	return entries, nil
}

// ReflogUndo resets HEAD to the given reflog entry (mixed reset).
func (c *Client) ReflogUndo(ref string) error {
	_, err := c.run("reset", "--mixed", ref)
	return err
}

func parseReflogSubject(s string) (action, subject string) {
	// Common patterns: "commit: message", "checkout: from X to Y", "rebase -i (finish)"
	knownActions := []string{
		"commit (amend)", "commit (merge)", "commit (cherry-pick)", "commit",
		"checkout", "rebase", "merge", "reset", "pull", "push", "clone",
		"branch", "stash",
	}
	for _, a := range knownActions {
		if strings.HasPrefix(s, a+": ") {
			return a, strings.TrimPrefix(s, a+": ")
		}
		if s == a {
			return a, ""
		}
	}
	return "op", s
}
