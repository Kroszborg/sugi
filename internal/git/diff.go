package git

import (
	"fmt"
	"strconv"
	"strings"
)

// DiffLineType describes the kind of change a diff line represents.
type DiffLineType int

const (
	DiffContext DiffLineType = iota
	DiffAdded
	DiffRemoved
	DiffHunkHeader
	DiffFileHeader
)

// DiffLine is a single line in a diff.
type DiffLine struct {
	Type    DiffLineType
	Content string
	OldLine int // line number in original file (0 = not applicable)
	NewLine int // line number in new file (0 = not applicable)
}

// DiffHunk is a contiguous block of changes in a diff.
type DiffHunk struct {
	Header  string // e.g. "@@ -1,5 +1,7 @@ func foo()"
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines   []DiffLine
}

// FileDiff contains all hunks for a single file.
type FileDiff struct {
	OldPath string
	NewPath string
	Hunks   []DiffHunk
	IsNew   bool
	IsDeleted bool
	IsBinary bool
}

// Diff returns the diff for a specific file.
// If staged is true, shows staged changes (--cached). Otherwise working-tree changes.
func (c *Client) Diff(path string, staged bool) ([]DiffHunk, error) {
	args := []string{"diff", "--unified=3"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", path)

	out, err := c.run(args...)
	if err != nil {
		return nil, err
	}
	fd := parseDiff(out)
	if len(fd) == 0 {
		return nil, nil
	}
	return fd[0].Hunks, nil
}

// DiffAll returns the full diff for all changed files.
func (c *Client) DiffAll(staged bool) ([]FileDiff, error) {
	args := []string{"diff", "--unified=3"}
	if staged {
		args = append(args, "--cached")
	}
	out, err := c.run(args...)
	if err != nil {
		return nil, err
	}
	return parseDiff(out), nil
}

// DiffCommit returns the diff introduced by a specific commit.
func (c *Client) DiffCommit(hash string) ([]FileDiff, error) {
	out, err := c.run("diff", "--unified=3", fmt.Sprintf("%s^", hash), hash)
	if err != nil {
		// Root commit has no parent
		out, err = c.run("show", "--unified=3", "--format=", hash)
		if err != nil {
			return nil, err
		}
	}
	return parseDiff(out), nil
}

// parseDiff parses unified diff output into FileDiff structs.
func parseDiff(raw string) []FileDiff {
	var files []FileDiff
	var current *FileDiff
	var currentHunk *DiffHunk
	oldLine, newLine := 0, 0

	for _, line := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git"):
			if current != nil {
				files = append(files, *current)
			}
			current = &FileDiff{}
			currentHunk = nil

		case current == nil:
			// Skip leading content before first diff header
			continue

		case strings.HasPrefix(line, "--- a/"):
			current.OldPath = strings.TrimPrefix(line, "--- a/")
		case strings.HasPrefix(line, "--- /dev/null"):
			current.IsNew = true
		case strings.HasPrefix(line, "+++ b/"):
			current.NewPath = strings.TrimPrefix(line, "+++ b/")
		case strings.HasPrefix(line, "+++ /dev/null"):
			current.IsDeleted = true
		case strings.HasPrefix(line, "Binary files"):
			current.IsBinary = true

		case strings.HasPrefix(line, "@@ "):
			if currentHunk != nil {
				current.Hunks = append(current.Hunks, *currentHunk)
			}
			hunk := parseHunkHeader(line)
			currentHunk = &hunk
			oldLine = hunk.OldStart
			newLine = hunk.NewStart
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:    DiffHunkHeader,
				Content: line,
			})

		case currentHunk != nil:
			switch {
			case strings.HasPrefix(line, "+"):
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    DiffAdded,
					Content: line[1:],
					NewLine: newLine,
				})
				newLine++
			case strings.HasPrefix(line, "-"):
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    DiffRemoved,
					Content: line[1:],
					OldLine: oldLine,
				})
				oldLine++
			case strings.HasPrefix(line, " "):
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    DiffContext,
					Content: line[1:],
					OldLine: oldLine,
					NewLine: newLine,
				})
				oldLine++
				newLine++
			case line == "\\ No newline at end of file":
				// skip
			}
		}
	}

	if currentHunk != nil && current != nil {
		current.Hunks = append(current.Hunks, *currentHunk)
	}
	if current != nil {
		files = append(files, *current)
	}
	return files
}

// parseHunkHeader parses "@@ -l,s +l,s @@ optional context" into a DiffHunk.
func parseHunkHeader(header string) DiffHunk {
	var hunk DiffHunk
	hunk.Header = header

	// Extract the @@ -a,b +c,d @@ part
	parts := strings.SplitN(header, " ", 4)
	if len(parts) < 3 {
		return hunk
	}

	parseRange := func(s string) (int, int) {
		s = strings.TrimPrefix(s, "-")
		s = strings.TrimPrefix(s, "+")
		idx := strings.Index(s, ",")
		if idx < 0 {
			n, _ := strconv.Atoi(s)
			return n, 1
		}
		start, _ := strconv.Atoi(s[:idx])
		count, _ := strconv.Atoi(s[idx+1:])
		return start, count
	}

	hunk.OldStart, hunk.OldCount = parseRange(parts[1])
	hunk.NewStart, hunk.NewCount = parseRange(parts[2])
	return hunk
}
