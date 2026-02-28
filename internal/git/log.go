package git

import (
	"strconv"
	"strings"
	"time"
)

// Commit represents a single git commit.
type Commit struct {
	Hash        string
	ShortHash   string
	Author      string
	AuthorEmail string
	Date        time.Time
	Subject     string
	Body        string
	Refs        []string // branch names and tags pointing here
}

// Log returns the commit history for the current branch.
func (c *Client) Log(limit int) ([]Commit, error) {
	sep := "\x1f" // unit separator
	format := strings.Join([]string{
		"%H",  // full hash
		"%h",  // short hash
		"%an", // author name
		"%ae", // author email
		"%ai", // author date ISO
		"%s",  // subject
		"%D",  // ref names
	}, sep)

	args := []string{
		"log",
		"--format=" + format,
	}
	if limit > 0 {
		args = append(args, "--max-count", strconv.Itoa(limit))
	}

	lines, err := c.runLines(args...)
	if err != nil {
		return nil, err
	}

	commits := make([]Commit, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, sep)
		if len(parts) < 6 {
			continue
		}

		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[4])

		var refs []string
		if len(parts) > 6 && parts[6] != "" {
			for _, ref := range strings.Split(parts[6], ", ") {
				ref = strings.TrimSpace(ref)
				if ref != "" {
					refs = append(refs, ref)
				}
			}
		}

		commits = append(commits, Commit{
			Hash:        parts[0],
			ShortHash:   parts[1],
			Author:      parts[2],
			AuthorEmail: parts[3],
			Date:        date,
			Subject:     parts[5],
			Refs:        refs,
		})
	}
	return commits, nil
}

// CommitDetail returns a single commit's full info including body.
func (c *Client) CommitDetail(hash string) (*Commit, error) {
	sep := "\x1f"
	format := strings.Join([]string{
		"%H", "%h", "%an", "%ae", "%ai", "%s", "%b", "%D",
	}, sep)

	out, err := c.run("show", "--format="+format, "--no-patch", hash)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(out, sep)
	if len(parts) < 6 {
		return nil, nil
	}

	date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[4])
	body := ""
	if len(parts) > 6 {
		body = strings.TrimSpace(parts[6])
	}

	var refs []string
	if len(parts) > 7 && parts[7] != "" {
		for _, ref := range strings.Split(parts[7], ", ") {
			ref = strings.TrimSpace(ref)
			if ref != "" {
				refs = append(refs, ref)
			}
		}
	}

	return &Commit{
		Hash:        parts[0],
		ShortHash:   parts[1],
		Author:      parts[2],
		AuthorEmail: parts[3],
		Date:        date,
		Subject:     parts[5],
		Body:        body,
		Refs:        refs,
	}, nil
}

// LogFile returns commits that touched a specific file path.
func (c *Client) LogFile(path string, limit int) ([]Commit, error) {
	sep := "\x1f"
	format := strings.Join([]string{
		"%H", "%h", "%an", "%ae", "%ai", "%s", "%D",
	}, sep)

	args := []string{
		"log",
		"--format=" + format,
		"--follow",
		"--",
		path,
	}
	if limit > 0 {
		args = []string{"log", "--format=" + format, "--follow", "--max-count", strconv.Itoa(limit), "--", path}
	}

	lines, err := c.runLines(args...)
	if err != nil {
		return nil, err
	}

	commits := make([]Commit, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, sep)
		if len(parts) < 6 {
			continue
		}
		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[4])
		var refs []string
		if len(parts) > 6 && parts[6] != "" {
			for _, ref := range strings.Split(parts[6], ", ") {
				ref = strings.TrimSpace(ref)
				if ref != "" {
					refs = append(refs, ref)
				}
			}
		}
		commits = append(commits, Commit{
			Hash:        parts[0],
			ShortHash:   parts[1],
			Author:      parts[2],
			AuthorEmail: parts[3],
			Date:        date,
			Subject:     parts[5],
			Refs:        refs,
		})
	}
	return commits, nil
}

// LogGraph returns formatted commit graph output for rendering.
func (c *Client) LogGraph(limit int) ([]string, error) {
	args := []string{
		"log",
		"--graph",
		"--pretty=format:%h %s",
		"--abbrev-commit",
		"--all",
	}
	if limit > 0 {
		args = append(args, "--max-count", strconv.Itoa(limit))
	}
	return c.runLines(args...)
}
