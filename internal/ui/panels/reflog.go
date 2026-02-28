package panels

import (
	"fmt"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// ReflogModel displays the git reflog (undo history).
type ReflogModel struct {
	entries []git.ReflogEntry
	list    widgets.ScrollList
	Width   int
	Height  int
}

// NewReflogModel creates a ReflogModel.
func NewReflogModel(width, height int) ReflogModel {
	return ReflogModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-2, width-4),
	}
}

// SetEntries updates the reflog list.
func (m *ReflogModel) SetEntries(entries []git.ReflogEntry) {
	m.entries = entries
	m.list.SetItems(m.buildItems())
}

// CurrentEntry returns the entry at the cursor.
func (m *ReflogModel) CurrentEntry() *git.ReflogEntry {
	if len(m.entries) == 0 || m.list.Cursor >= len(m.entries) {
		return nil
	}
	return &m.entries[m.list.Cursor]
}

// MoveUp moves the cursor up.
func (m *ReflogModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *ReflogModel) MoveDown() { m.list.MoveDown() }

// View renders the reflog panel.
func (m *ReflogModel) View() string {
	if len(m.entries) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render("  No reflog entries")
	}
	return m.list.View()
}

func (m *ReflogModel) buildItems() []string {
	items := make([]string, len(m.entries))
	for i, e := range m.entries {
		ref := lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(e.Ref)
		hash := lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true).Render(e.ShortHash)
		action := actionStyle(e.Action).Render(e.Action)
		subject := lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Render(e.Subject)
		date := lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(relativeDate(e.Date))
		items[i] = fmt.Sprintf(" %s %s %s %s  %s", ref, hash, action, subject, date)
	}
	return items
}

func actionStyle(action string) lipgloss.Style {
	switch action {
	case "commit", "commit (amend)", "commit (merge)", "commit (cherry-pick)":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true)
	case "checkout":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))
	case "reset":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))
	case "merge":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7"))
	case "rebase":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8"))
	}
}

// ListCursor returns the current scroll list cursor position.
func (m *ReflogModel) ListCursor() int { return m.list.Cursor }

// SetListCursor sets the scroll list cursor position.
func (m *ReflogModel) SetListCursor(n int) { m.list.Cursor = n }
