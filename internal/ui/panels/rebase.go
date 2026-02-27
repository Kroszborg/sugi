package panels

import (
	"fmt"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// RebaseModel displays the interactive rebase todo list.
type RebaseModel struct {
	Entries  []git.RebaseTodoEntry
	TodoPath string
	list     widgets.ScrollList
	Width    int
	Height   int
}

// NewRebaseModel creates a RebaseModel.
func NewRebaseModel(width, height int) RebaseModel {
	return RebaseModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-2, width-4),
	}
}

// SetEntries updates the rebase todo entries.
func (m *RebaseModel) SetEntries(entries []git.RebaseTodoEntry, todoPath string) {
	m.Entries = entries
	m.TodoPath = todoPath
	m.list.SetItems(m.buildItems())
}

// CurrentIndex returns the cursor position.
func (m *RebaseModel) CurrentIndex() int { return m.list.Cursor }

// CycleAction cycles the action on the current entry.
func (m *RebaseModel) CycleAction() {
	i := m.list.Cursor
	if i < len(m.Entries) {
		m.Entries[i].CycleAction()
		m.list.SetItems(m.buildItems())
	}
}

// MoveEntryUp moves the current entry up (swaps with previous).
func (m *RebaseModel) MoveEntryUp() {
	i := m.list.Cursor
	if i > 0 {
		m.Entries[i-1], m.Entries[i] = m.Entries[i], m.Entries[i-1]
		m.list.MoveUp()
		m.list.SetItems(m.buildItems())
	}
}

// MoveEntryDown moves the current entry down (swaps with next).
func (m *RebaseModel) MoveEntryDown() {
	i := m.list.Cursor
	if i < len(m.Entries)-1 {
		m.Entries[i+1], m.Entries[i] = m.Entries[i], m.Entries[i+1]
		m.list.MoveDown()
		m.list.SetItems(m.buildItems())
	}
}

// MoveUp moves the cursor up.
func (m *RebaseModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *RebaseModel) MoveDown() { m.list.MoveDown() }

// View renders the rebase todo panel.
func (m *RebaseModel) View() string {
	if len(m.Entries) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render("  No rebase entries")
	}
	return m.list.View()
}

func (m *RebaseModel) buildItems() []string {
	items := make([]string, len(m.Entries))
	for i, e := range m.Entries {
		items[i] = renderRebaseEntry(e)
	}
	return items
}

func renderRebaseEntry(e git.RebaseTodoEntry) string {
	var actionStyle lipgloss.Style
	switch e.Action {
	case git.RebasePick:
		actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true)
	case git.RebaseReword:
		actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Bold(true)
	case git.RebaseSquash:
		actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true)
	case git.RebaseFixup:
		actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true)
	case git.RebaseDrop:
		actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Bold(true)
	default:
		actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70"))
	}

	action := actionStyle.Render(fmt.Sprintf("%-6s", string(e.Action)))
	hash := lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Render(e.Hash)
	subject := lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Render(e.Subject)

	return fmt.Sprintf(" %s %s  %s", action, hash, subject)
}
