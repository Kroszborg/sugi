package panels

import (
	"fmt"
	"time"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// CommitModel displays the commit history.
type CommitModel struct {
	commits []git.Commit
	list    widgets.ScrollList
	Width   int
	Height  int
}

// NewCommitModel creates a CommitModel with the given dimensions.
func NewCommitModel(width, height int) CommitModel {
	return CommitModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-2, width-4),
	}
}

// SetCommits updates the commit list.
func (m *CommitModel) SetCommits(commits []git.Commit) {
	m.commits = commits
	m.list.SetItems(m.buildItems())
}

// CurrentCommit returns the commit at the cursor, or nil.
func (m *CommitModel) CurrentCommit() *git.Commit {
	if len(m.commits) == 0 || m.list.Cursor >= len(m.commits) {
		return nil
	}
	return &m.commits[m.list.Cursor]
}

// MoveUp moves the cursor up.
func (m *CommitModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *CommitModel) MoveDown() { m.list.MoveDown() }

// View renders the commits panel content.
func (m *CommitModel) View() string {
	if len(m.commits) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render("  No commits")
	}
	return m.list.View()
}

func (m *CommitModel) buildItems() []string {
	items := make([]string, len(m.commits))
	for i, c := range m.commits {
		items[i] = renderCommitItem(c, m.Width-4)
	}
	return items
}

func renderCommitItem(c git.Commit, width int) string {
	hash := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fab387")).
		Bold(true).
		Render(c.ShortHash)

	refs := ""
	for _, ref := range c.Refs {
		refs += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f9e2af")).
			Render(" [" + ref + "]")
	}

	subject := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cdd6f4")).
		Render(c.Subject)

	date := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#585b70")).
		Render(relativeDate(c.Date))

	// Truncate subject if needed
	hashW := 8
	dateW := 8
	subjectW := width - hashW - dateW - len(refs) - 4
	if subjectW < 10 {
		subjectW = 10
	}
	sub := c.Subject
	if len(sub) > subjectW {
		sub = sub[:subjectW-1] + "…"
	}
	subject = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Render(sub)

	return fmt.Sprintf(" %s %s%s  %s", hash, subject, refs, date)
}

func relativeDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy ago", int(d.Hours()/(24*365)))
	}
}
