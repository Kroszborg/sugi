package panels

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/charmbracelet/lipgloss"
)

// DiffModel displays a unified diff with syntax highlighting.
type DiffModel struct {
	hunks  []git.DiffHunk
	lines  []string // rendered lines
	offset int      // vertical scroll offset
	Width  int
	Height int

	// Styles
	addedStyle   lipgloss.Style
	removedStyle lipgloss.Style
	contextStyle lipgloss.Style
	hunkStyle    lipgloss.Style
	headerStyle  lipgloss.Style
	emptyStyle   lipgloss.Style
}

// NewDiffModel creates a DiffModel with the given dimensions.
func NewDiffModel(width, height int) DiffModel {
	return DiffModel{
		Width:        width,
		Height:       height,
		addedStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
		removedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
		contextStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
		hunkStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#94e2d5")),
		headerStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true),
		emptyStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")),
	}
}

// SetHunks updates the diff content.
func (m *DiffModel) SetHunks(hunks []git.DiffHunk) {
	m.hunks = hunks
	m.offset = 0
	m.lines = m.buildLines()
}

// SetFileDiff sets the diff from a FileDiff struct.
func (m *DiffModel) SetFileDiff(fd *git.FileDiff) {
	if fd == nil {
		m.hunks = nil
		m.lines = nil
		m.offset = 0
		return
	}
	m.SetHunks(fd.Hunks)
}

// Clear empties the diff view.
func (m *DiffModel) Clear() {
	m.hunks = nil
	m.lines = nil
	m.offset = 0
}

// ScrollUp scrolls up by one line.
func (m *DiffModel) ScrollUp() {
	if m.offset > 0 {
		m.offset--
	}
}

// ScrollDown scrolls down by one line.
func (m *DiffModel) ScrollDown() {
	maxOffset := len(m.lines) - m.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset < maxOffset {
		m.offset++
	}
}

// PageUp scrolls up by one page.
func (m *DiffModel) PageUp() {
	m.offset -= m.Height
	if m.offset < 0 {
		m.offset = 0
	}
}

// PageDown scrolls down by one page.
func (m *DiffModel) PageDown() {
	m.offset += m.Height
	maxOffset := len(m.lines) - m.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

// View renders the diff panel content.
func (m *DiffModel) View() string {
	if len(m.lines) == 0 {
		return m.emptyStyle.Render("  Select a file to view diff")
	}

	end := m.offset + m.Height
	if end > len(m.lines) {
		end = len(m.lines)
	}

	visible := m.lines[m.offset:end]
	return strings.Join(visible, "\n")
}

// ScrollInfo returns "offset/total" for the status bar.
func (m *DiffModel) ScrollInfo() string {
	if len(m.lines) == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", m.offset+1, len(m.lines))
}

func (m *DiffModel) buildLines() []string {
	var lines []string
	w := m.Width - 4
	if w < 10 {
		w = 10
	}

	for _, hunk := range m.hunks {
		for _, dl := range hunk.Lines {
			line := m.renderLine(dl, w)
			lines = append(lines, line)
		}
	}
	return lines
}

func (m *DiffModel) renderLine(dl git.DiffLine, width int) string {
	content := dl.Content
	if len(content) > width-2 {
		content = content[:width-3] + "…"
	}

	switch dl.Type {
	case git.DiffAdded:
		return m.addedStyle.Render("+" + content)
	case git.DiffRemoved:
		return m.removedStyle.Render("-" + content)
	case git.DiffHunkHeader:
		return m.hunkStyle.Render(content)
	case git.DiffFileHeader:
		return m.headerStyle.Render(content)
	default:
		return m.contextStyle.Render(" " + content)
	}
}
