package panels

import (
	"fmt"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// WorktreeModel displays the list of git worktrees.
type WorktreeModel struct {
	worktrees []git.Worktree
	list      widgets.ScrollList
	Width     int
	Height    int
}

// NewWorktreeModel creates a WorktreeModel.
func NewWorktreeModel(width, height int) WorktreeModel {
	return WorktreeModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-3, width-4),
	}
}

// SetWorktrees updates the worktree list.
func (m *WorktreeModel) SetWorktrees(wts []git.Worktree) {
	m.worktrees = wts
	m.list.SetItems(m.buildItems())
}

// CurrentWorktree returns the worktree at the cursor.
func (m *WorktreeModel) CurrentWorktree() *git.Worktree {
	if len(m.worktrees) == 0 || m.list.Cursor >= len(m.worktrees) {
		return nil
	}
	return &m.worktrees[m.list.Cursor]
}

// MoveUp moves the cursor up.
func (m *WorktreeModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *WorktreeModel) MoveDown() { m.list.MoveDown() }

// View renders the worktrees panel.
func (m *WorktreeModel) View() string {
	if len(m.worktrees) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).Render("  No worktrees")
	}
	return m.list.View()
}

func (m *WorktreeModel) buildItems() []string {
	items := make([]string, len(m.worktrees))
	for i, wt := range m.worktrees {
		items[i] = renderWorktreeItem(wt)
	}
	return items
}

func renderWorktreeItem(wt git.Worktree) string {
	icon := "  "
	if wt.IsMain {
		icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")).Render("● ")
	} else if wt.IsLocked {
		icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#e05454")).Render("🔒 ")
	}

	path := lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee")).Render(wt.Path)

	branch := ""
	if wt.Branch != "" {
		branch = lipgloss.NewStyle().Foreground(lipgloss.Color("#4d9de0")).Render(fmt.Sprintf(" [%s]", wt.Branch))
	}

	head := ""
	if wt.Head != "" {
		short := wt.Head
		if len(short) > 7 {
			short = short[:7]
		}
		head = lipgloss.NewStyle().Foreground(lipgloss.Color("#e8835c")).Render(" " + short)
	}

	extra := ""
	if wt.IsBare {
		extra = lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).Render(" (bare)")
	}

	return fmt.Sprintf(" %s%s%s%s%s", icon, path, branch, head, extra)
}

// ListCursor returns the current scroll list cursor position.
func (m *WorktreeModel) ListCursor() int { return m.list.Cursor }

// SetListCursor sets the scroll list cursor position.
func (m *WorktreeModel) SetListCursor(n int) { m.list.Cursor = n }

// GetWorktrees returns the raw worktree slice (for rebuildPanels preservation).
func (m *WorktreeModel) GetWorktrees() []git.Worktree { return m.worktrees }
