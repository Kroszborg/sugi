package panels

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// ConflictModel displays merge conflict blocks for a file.
type ConflictModel struct {
	FilePath string
	Blocks   []git.ConflictBlock
	list     widgets.ScrollList
	Width    int
	Height   int
}

// NewConflictModel creates a ConflictModel.
func NewConflictModel(width, height int) ConflictModel {
	return ConflictModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-3, width-4),
	}
}

// SetConflicts updates the conflict blocks.
func (m *ConflictModel) SetConflicts(path string, blocks []git.ConflictBlock) {
	m.FilePath = path
	m.Blocks = blocks
	m.list.SetItems(m.buildItems())
}

// CurrentBlock returns the conflict block at the cursor.
func (m *ConflictModel) CurrentBlock() *git.ConflictBlock {
	if len(m.Blocks) == 0 || m.list.Cursor >= len(m.Blocks) {
		return nil
	}
	return &m.Blocks[m.list.Cursor]
}

// MoveUp moves the cursor up.
func (m *ConflictModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *ConflictModel) MoveDown() { m.list.MoveDown() }

// View renders the conflict panel.
func (m *ConflictModel) View() string {
	if len(m.Blocks) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")).Render("  No conflicts — file is clean")
	}
	return m.list.View()
}

func (m *ConflictModel) buildItems() []string {
	items := make([]string, len(m.Blocks))
	for i, b := range m.Blocks {
		items[i] = renderConflictBlock(i, b)
	}
	return items
}

func renderConflictBlock(idx int, b git.ConflictBlock) string {
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d4a017")).Bold(true).
		Render(fmt.Sprintf("  Conflict #%d  (line %d)", idx+1, b.Start+1))

	oursLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")).Render("  ours:   ")
	theirsLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#4d9de0")).Render("  theirs: ")

	oursPreview := truncateLines(b.OursLines, 2)
	theirsPreview := truncateLines(b.TheirsLines, 2)

	return header + "\n" + oursLabel +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee")).Render(oursPreview) +
		"\n" + theirsLabel +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee")).Render(theirsPreview)
}

func truncateLines(lines []string, max int) string {
	if len(lines) == 0 {
		return "(empty)"
	}
	if len(lines) > max {
		lines = lines[:max]
	}
	return strings.Join(lines, " / ")
}

// ListCursor returns the current scroll list cursor position.
func (m *ConflictModel) ListCursor() int { return m.list.Cursor }

// SetListCursor sets the scroll list cursor position.
func (m *ConflictModel) SetListCursor(n int) { m.list.Cursor = n }
