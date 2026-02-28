package git

import "strings"

// StatusCode represents a single-character git status code.
type StatusCode rune

const (
	Modified  StatusCode = 'M'
	Added     StatusCode = 'A'
	Deleted   StatusCode = 'D'
	Renamed   StatusCode = 'R'
	Copied    StatusCode = 'C'
	Untracked StatusCode = '?'
	Unmerged  StatusCode = 'U'
	Ignored   StatusCode = '!'
	None      StatusCode = ' '
)

// FileStatus represents a file's status in the working tree.
type FileStatus struct {
	Staged   StatusCode // X - index status
	Unstaged StatusCode // Y - worktree status
	Path     string
	OldPath  string // populated for renamed files
}

// IsStaged returns true if the file has staged changes.
func (f FileStatus) IsStaged() bool {
	return f.Staged != None && f.Staged != Untracked && f.Staged != Ignored
}

// IsUnstaged returns true if the file has unstaged changes.
func (f FileStatus) IsUnstaged() bool {
	return f.Unstaged != None && f.Unstaged != Untracked && f.Unstaged != Ignored
}

// IsUntracked returns true if the file is untracked.
func (f FileStatus) IsUntracked() bool {
	return f.Staged == Untracked && f.Unstaged == Untracked
}

// IsConflicted returns true if the file has merge conflicts.
func (f FileStatus) IsConflicted() bool {
	// Conflict codes: DD, AU, UD, UA, DU, AA, UU
	return f.Staged == Unmerged || f.Unstaged == Unmerged ||
		(f.Staged == Added && f.Unstaged == Added) ||
		(f.Staged == Deleted && f.Unstaged == Deleted)
}

// Status returns all modified/staged/untracked files using porcelain v1 format.
func (c *Client) Status() ([]FileStatus, error) {
	out, err := c.run("status", "--porcelain=v1")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	lines := strings.Split(out, "\n")
	files := make([]FileStatus, 0, len(lines))
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		xy := line[:2]
		path := line[3:]

		fs := FileStatus{
			Staged:   StatusCode(xy[0]),
			Unstaged: StatusCode(xy[1]),
		}

		// Renamed format: "R  newpath -> oldpath" or "R  newpath\toldpath"
		if xy[0] == 'R' || xy[0] == 'C' {
			// git porcelain v1: "R  old -> new" is not actually the format
			// The format for renames in v1 is "R  newfile\x00oldfile" but
			// only in -z mode. Without -z it's "R  new -> old"
			parts := strings.SplitN(path, " -> ", 2)
			if len(parts) == 2 {
				fs.Path = parts[0]
				fs.OldPath = parts[1]
			} else {
				fs.Path = path
			}
		} else {
			fs.Path = path
		}

		files = append(files, fs)
	}
	return files, nil
}
