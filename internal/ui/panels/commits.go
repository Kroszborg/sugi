package panels

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// CommitModel displays the commit history with optional graph.
type CommitModel struct {
	commits     []git.Commit
	graphLines  []string // raw git log --graph lines
	list        widgets.ScrollList
	Width       int
	Height      int
	ShowGraph   bool
	graphRender widgets.GraphRenderer

	searchFilter string
	filtered     []int // indices into commits that match filter
}

// NewCommitModel creates a CommitModel with the given dimensions.
func NewCommitModel(width, height int) CommitModel {
	return CommitModel{
		Width:       width,
		Height:      height,
		list:        widgets.NewScrollList(height-3, width-4),
		ShowGraph:   false,
		graphRender: widgets.NewGraphRenderer(),
	}
}

// SetCommits updates the commit list.
func (m *CommitModel) SetCommits(commits []git.Commit) {
	m.commits = commits
	m.searchFilter = ""
	m.filtered = nil
	m.list.SetItems(m.buildItems())
}

// SetGraphLines sets raw git log --graph output for graph mode.
func (m *CommitModel) SetGraphLines(lines []string) {
	m.graphLines = lines
	// If graph is already toggled on, render immediately now that data arrived.
	if m.ShowGraph && len(lines) > 0 {
		m.list.SetItems(m.graphRender.RenderLines(lines, m.Width-4))
	}
}

// ToggleGraph toggles the commit graph view.
func (m *CommitModel) ToggleGraph() {
	m.ShowGraph = !m.ShowGraph
	if m.ShowGraph && len(m.graphLines) > 0 {
		m.list.SetItems(m.graphRender.RenderLines(m.graphLines, m.Width-4))
	} else {
		m.list.SetItems(m.buildItems())
	}
}

// SetFilter applies a search filter to the commit list.
func (m *CommitModel) SetFilter(q string) {
	m.searchFilter = q
	if q == "" {
		m.filtered = nil
		m.list.SetItems(m.buildItems())
		return
	}
	q = strings.ToLower(q)
	m.filtered = nil
	var items []string
	for i, c := range m.commits {
		if strings.Contains(strings.ToLower(c.Subject), q) ||
			strings.Contains(strings.ToLower(c.Author), q) ||
			strings.HasPrefix(c.ShortHash, q) {
			m.filtered = append(m.filtered, i)
			items = append(items, renderCommitItem(c, m.Width-4))
		}
	}
	m.list.SetItems(items)
}

// CurrentCommit returns the commit at the cursor, respecting filters.
func (m *CommitModel) CurrentCommit() *git.Commit {
	if len(m.commits) == 0 {
		return nil
	}
	if m.ShowGraph {
		// In graph mode, cursor maps to graph lines, not 1:1 with commits
		return nil
	}
	idx := m.list.Cursor
	if m.filtered != nil {
		if idx >= len(m.filtered) {
			return nil
		}
		idx = m.filtered[idx]
	}
	if idx >= len(m.commits) {
		return nil
	}
	return &m.commits[idx]
}

// MoveUp moves the cursor up.
func (m *CommitModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *CommitModel) MoveDown() { m.list.MoveDown() }

// PageUp pages up.
func (m *CommitModel) PageUp() { m.list.PageUp() }

// PageDown pages down.
func (m *CommitModel) PageDown() { m.list.PageDown() }

// View renders the commits panel content.
func (m *CommitModel) View() string {
	if len(m.commits) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).Render("  No commits")
	}
	if m.searchFilter != "" && len(m.filtered) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).
			Render(fmt.Sprintf("  No commits matching %q", m.searchFilter))
	}
	return m.list.View()
}

// Count returns the number of visible commits.
func (m *CommitModel) Count() int {
	if m.filtered != nil {
		return len(m.filtered)
	}
	return len(m.commits)
}

func (m *CommitModel) buildItems() []string {
	src := m.commits
	items := make([]string, len(src))
	for i, c := range src {
		items[i] = renderCommitItem(c, m.Width-4)
	}
	return items
}

func renderCommitItem(c git.Commit, width int) string {
	hash := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e8835c")).
		Bold(true).
		Render(c.ShortHash)

	refs := ""
	for _, ref := range c.Refs {
		// Skip HEAD ->
		if strings.HasPrefix(ref, "HEAD -> ") {
			ref = strings.TrimPrefix(ref, "HEAD -> ")
			refs += lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")).Bold(true).Render(" [" + ref + "]")
		} else if strings.HasPrefix(ref, "tag: ") {
			refs += lipgloss.NewStyle().Foreground(lipgloss.Color("#d4a017")).Render(" [" + ref + "]")
		} else {
			refs += lipgloss.NewStyle().Foreground(lipgloss.Color("#a87efb")).Render(" [" + ref + "]")
		}
	}

	date := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3d3d5c")).
		Render(relativeDate(c.Date))

	// Calculate available subject width
	hashW := 8
	dateW := len(relativeDate(c.Date)) + 2
	refsW := lipgloss.Width(refs)
	subjectW := width - hashW - dateW - refsW - 4
	if subjectW < 10 {
		subjectW = 10
	}

	sub := c.Subject
	if len(sub) > subjectW {
		sub = sub[:subjectW-1] + "…"
	}
	subject := lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee")).Render(sub)

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
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy", int(d.Hours()/(24*365)))
	}
}

// ListCursor returns the current scroll list cursor position.
func (m *CommitModel) ListCursor() int { return m.list.Cursor }

// SetListCursor sets the scroll list cursor position.
func (m *CommitModel) SetListCursor(n int) { m.list.Cursor = n }
