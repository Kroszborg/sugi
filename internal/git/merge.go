package git

import (
	"os"
	"strings"
)

// ConflictBlock represents a single conflict hunk in a file.
type ConflictBlock struct {
	OursLines   []string
	TheirsLines []string
	Start       int // line index of <<<<<<< marker
	End         int // line index of >>>>>>> marker
}

// ConflictedFile reads a file and returns its conflict blocks.
func ConflictedFile(path string) ([]ConflictBlock, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	lines := strings.Split(string(data), "\n")
	return parseConflicts(lines), lines, nil
}

func parseConflicts(lines []string) []ConflictBlock {
	var blocks []ConflictBlock
	i := 0
	for i < len(lines) {
		if strings.HasPrefix(lines[i], "<<<<<<<") {
			block := ConflictBlock{Start: i}
			i++
			// collect ours
			for i < len(lines) && !strings.HasPrefix(lines[i], "=======") {
				block.OursLines = append(block.OursLines, lines[i])
				i++
			}
			i++ // skip =======
			// collect theirs
			for i < len(lines) && !strings.HasPrefix(lines[i], ">>>>>>>") {
				block.TheirsLines = append(block.TheirsLines, lines[i])
				i++
			}
			block.End = i
			blocks = append(blocks, block)
		}
		i++
	}
	return blocks
}

// ResolveConflict resolves a conflict in a file by choosing "ours" or "theirs".
// side must be "ours" or "theirs".
func ResolveConflict(path, side string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	var out []string
	i := 0
	for i < len(lines) {
		if strings.HasPrefix(lines[i], "<<<<<<<") {
			i++
			var ours, theirs []string
			for i < len(lines) && !strings.HasPrefix(lines[i], "=======") {
				ours = append(ours, lines[i])
				i++
			}
			i++ // skip =======
			for i < len(lines) && !strings.HasPrefix(lines[i], ">>>>>>>") {
				theirs = append(theirs, lines[i])
				i++
			}
			if side == "ours" {
				out = append(out, ours...)
			} else {
				out = append(out, theirs...)
			}
		} else {
			out = append(out, lines[i])
		}
		i++
	}
	return os.WriteFile(path, []byte(strings.Join(out, "\n")), 0644)
}

// MarkResolved stages the file (marks it resolved in git index).
func (c *Client) MarkResolved(path string) error {
	return c.Stage(path)
}
