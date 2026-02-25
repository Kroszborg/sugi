package panels

import (
	"fmt"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// StashModel displays the stash list with diff preview.
type StashModel struct {
	stashes []git.StashEntry
	list    widgets.ScrollList
	diff    DiffModel
	Width   int
	Height  int

	showDiff bool
}

// NewStashModel creates a StashModel.
func NewStashModel(width, height int) StashModel {
	listH := height / 2
	diffH := height - listH
	return StashModel{
		Width:    width,
		Height:   height,
		list:     widgets.NewScrollList(listH-2, width-4),
		diff:     NewDiffModel(width, diffH),
		showDiff: true,
	}
}

// SetStashes updates the stash list.
func (m *StashModel) SetStashes(stashes []git.StashEntry) {
	m.stashes = stashes
	m.list.SetItems(m.buildItems())
}

// SetDiff sets the diff preview for the selected stash.
func (m *StashModel) SetDiff(fds []git.FileDiff) {
	var allHunks []git.DiffHunk
	for _, fd := range fds {
		allHunks = append(allHunks, fd.Hunks...)
	}
	m.diff.SetHunks(allHunks)
}

// CurrentStash returns the selected stash entry.
func (m *StashModel) CurrentStash() *git.StashEntry {
	if len(m.stashes) == 0 || m.list.Cursor >= len(m.stashes) {
		return nil
	}
	return &m.stashes[m.list.Cursor]
}

// MoveUp moves the cursor up.
func (m *StashModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *StashModel) MoveDown() { m.list.MoveDown() }

// ScrollDiffUp scrolls the diff preview up.
func (m *StashModel) ScrollDiffUp() { m.diff.ScrollUp() }

// ScrollDiffDown scrolls the diff preview down.
func (m *StashModel) ScrollDiffDown() { m.diff.ScrollDown() }

// View renders the stash panel (list + diff preview).
func (m *StashModel) View() string {
	listView := m.list.View()

	if !m.showDiff || len(m.stashes) == 0 {
		return listView
	}

	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#45475a")).
		Render(fmt.Sprintf("─── diff preview %s", "─────────────────────"))

	return listView + "\n" + sep + "\n" + m.diff.View()
}

func (m *StashModel) buildItems() []string {
	items := make([]string, len(m.stashes))
	for i, s := range m.stashes {
		date := ""
		if !s.Date.IsZero() {
			date = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(relativeDate(s.Date))
		}
		branch := ""
		if s.Branch != "" {
			branch = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Render(" [" + s.Branch + "]")
		}
		idx := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true).Render(fmt.Sprintf("stash@{%d}", i))
		msg := lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Render(s.Message)
		items[i] = fmt.Sprintf(" %s%s %s  %s", idx, branch, msg, date)
	}
	return items
}
