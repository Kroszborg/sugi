package git

import (
	"strconv"
	"strings"
	"time"
)

// BlameLine is a single line of git blame output.
type BlameLine struct {
	Hash      string
	ShortHash string
	Author    string
	Date      time.Time
	LineNum   int
	Content   string
	IsFirst   bool // first line attributed to this commit in the block
}

// Blame returns annotated blame information for a file.
func (c *Client) Blame(path string) ([]BlameLine, error) {
	// Use porcelain format for reliable parsing
	out, err := c.run("blame", "--porcelain", "--", path)
	if err != nil {
		return nil, err
	}
	return parseBlamePorcelain(out), nil
}

// BlameAt returns blame for a specific commit version of a file.
func (c *Client) BlameAt(hash, path string) ([]BlameLine, error) {
	out, err := c.run("blame", "--porcelain", hash, "--", path)
	if err != nil {
		return nil, err
	}
	return parseBlamePorcelain(out), nil
}

// parseBlamePorcelain parses git blame --porcelain output.
// Format: each commit block starts with "<hash> <orig-line> <final-line> [<num-lines>]"
// followed by header fields, then "\t<content>".
func parseBlamePorcelain(raw string) []BlameLine {
	type commitInfo struct {
		hash   string
		author string
		date   time.Time
		seen   bool
	}

	commits := make(map[string]*commitInfo)
	var lines []BlameLine

	rawLines := strings.Split(raw, "\n")
	i := 0
	for i < len(rawLines) {
		line := rawLines[i]

		// Header line: "<40-char-hash> <orig-line> <final-line> [count]"
		parts := strings.Fields(line)
		if len(parts) < 3 || len(parts[0]) != 40 {
			i++
			continue
		}

		hash := parts[0]
		ci, exists := commits[hash]
		if !exists {
			ci = &commitInfo{hash: hash}
			commits[hash] = ci
		}
		isFirst := !ci.seen

		// Read commit metadata lines until we hit the content line (starts with \t)
		i++
		for i < len(rawLines) && !strings.HasPrefix(rawLines[i], "\t") {
			field := rawLines[i]
			if strings.HasPrefix(field, "author ") {
				ci.author = strings.TrimPrefix(field, "author ")
			} else if strings.HasPrefix(field, "author-time ") {
				// Unix timestamp — use strconv to safely parse and avoid overflow
				ts := strings.TrimSpace(strings.TrimPrefix(field, "author-time "))
				if unix, err := strconv.ParseInt(ts, 10, 64); err == nil {
					ci.date = time.Unix(unix, 0)
				}
			}
			i++
		}

		// Content line (starts with \t)
		content := ""
		if i < len(rawLines) && strings.HasPrefix(rawLines[i], "\t") {
			content = rawLines[i][1:]
			i++
		}

		if !ci.seen {
			ci.seen = true
		}

		lineNum := len(lines) + 1
		lines = append(lines, BlameLine{
			Hash:      hash,
			ShortHash: hash[:8],
			Author:    ci.author,
			Date:      ci.date,
			LineNum:   lineNum,
			Content:   content,
			IsFirst:   isFirst,
		})
	}

	return lines
}
