package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RebaseTodoAction represents an action in a rebase todo list.
type RebaseTodoAction string

const (
	RebasePick   RebaseTodoAction = "pick"
	RebaseReword RebaseTodoAction = "reword"
	RebaseSquash RebaseTodoAction = "squash"
	RebaseFixup  RebaseTodoAction = "fixup"
	RebaseDrop   RebaseTodoAction = "drop"
)

// RebaseTodoEntry is one line in the interactive rebase todo file.
type RebaseTodoEntry struct {
	Action  RebaseTodoAction
	Hash    string
	Subject string
}

// CycleAction cycles through rebase actions in order.
func (e *RebaseTodoEntry) CycleAction() {
	switch e.Action {
	case RebasePick:
		e.Action = RebaseReword
	case RebaseReword:
		e.Action = RebaseSquash
	case RebaseSquash:
		e.Action = RebaseFixup
	case RebaseFixup:
		e.Action = RebaseDrop
	default:
		e.Action = RebasePick
	}
}

// RebaseInProgress returns true if a rebase is currently in progress.
func (c *Client) RebaseInProgress() bool {
	_, err := os.Stat(filepath.Join(c.RepoPath, ".git", "rebase-merge"))
	if err == nil {
		return true
	}
	_, err = os.Stat(filepath.Join(c.RepoPath, ".git", "rebase-apply"))
	return err == nil
}

// MergeInProgress returns true if a merge is currently in progress.
func (c *Client) MergeInProgress() bool {
	_, err := os.Stat(filepath.Join(c.RepoPath, ".git", "MERGE_HEAD"))
	return err == nil
}

// StartInteractiveRebase begins an interactive rebase using a passthrough
// sequence editor, then reads the generated todo file.
func (c *Client) StartInteractiveRebase(upstream string) ([]RebaseTodoEntry, string, error) {
	// Use "true" as GIT_SEQUENCE_EDITOR so git creates the rebase-merge dir
	// without opening an editor — we read and write the file ourselves.
	cmd := c.buildCmd("rebase", "-i", upstream)
	cmd.Env = append(os.Environ(), "GIT_SEQUENCE_EDITOR=true")
	_ = cmd.Run() // may "fail" since we intercept the editor step

	todoPath := filepath.Join(c.RepoPath, ".git", "rebase-merge", "git-rebase-todo")
	data, err := os.ReadFile(todoPath)
	if err != nil {
		return nil, "", fmt.Errorf("rebase todo not found: %w", err)
	}

	entries := parseRebaseTodo(string(data))
	return entries, todoPath, nil
}

// WriteRebaseTodo writes updated entries back to the todo file.
func WriteRebaseTodo(todoPath string, entries []RebaseTodoEntry) error {
	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("%s %s %s\n", e.Action, e.Hash, e.Subject))
	}
	return os.WriteFile(todoPath, []byte(sb.String()), 0644)
}

// ContinueRebase continues an in-progress rebase.
func (c *Client) ContinueRebase() error {
	cmd := c.buildCmd("rebase", "--continue")
	cmd.Env = append(os.Environ(), "GIT_EDITOR=true")
	return cmd.Run()
}

func parseRebaseTodo(raw string) []RebaseTodoEntry {
	var entries []RebaseTodoEntry
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 2 {
			continue
		}
		action := RebaseTodoAction(parts[0])
		hash := parts[1]
		subject := ""
		if len(parts) > 2 {
			subject = parts[2]
		}
		entries = append(entries, RebaseTodoEntry{
			Action:  action,
			Hash:    hash,
			Subject: subject,
		})
	}
	return entries
}
