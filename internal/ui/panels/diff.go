package panels

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/charmbracelet/lipgloss"
)

// DiffModel displays a unified diff with hunk-level navigation and staging.
type DiffModel struct {
	hunks      []git.DiffHunk
	lines      []string // rendered display lines
	lineToHunk []int    // maps display line index -> hunk index (-1 = non-hunk)
	hunkStarts []int    // display line where each hunk begins
	offset     int      // vertical scroll offset
	hunkCursor int      // which hunk is selected (for staging)
	Width      int
	Height     int

	addedStyle   lipgloss.Style
	removedStyle lipgloss.Style
	contextStyle lipgloss.Style
	hunkStyle    lipgloss.Style
	hunkFocused  lipgloss.Style
	headerStyle  lipgloss.Style
	emptyStyle   lipgloss.Style
}

// NewDiffModel creates a DiffModel with the given dimensions.
func NewDiffModel(width, height int) DiffModel {
	return DiffModel{
		Width:        width,
		Height:       height,
		addedStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")).Background(lipgloss.Color("#0e2018")),
		removedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#e05454")).Background(lipgloss.Color("#200e0e")),
		contextStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee")),
		hunkStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#2ec4b6")),
		hunkFocused:  lipgloss.NewStyle().Foreground(lipgloss.Color("#2ec4b6")).Background(lipgloss.Color("#1e3a4a")).Bold(true),
		headerStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#4d9de0")).Bold(true),
		emptyStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")),
	}
}

// SetHunks updates the diff content and rebuilds display lines.
func (m *DiffModel) SetHunks(hunks []git.DiffHunk) {
	m.hunks = hunks
	m.offset = 0
	m.hunkCursor = 0
	m.buildLines()
}

// SetFileDiff sets diff from a FileDiff struct.
func (m *DiffModel) SetFileDiff(fd *git.FileDiff) {
	if fd == nil {
		m.hunks = nil
		m.lines = nil
		m.lineToHunk = nil
		m.hunkStarts = nil
		m.offset = 0
		return
	}
	m.SetHunks(fd.Hunks)
}

// Clear empties the diff view.
func (m *DiffModel) Clear() {
	m.hunks = nil
	m.lines = nil
	m.lineToHunk = nil
	m.hunkStarts = nil
	m.offset = 0
	m.hunkCursor = 0
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

// NextHunk moves the hunk cursor to the next hunk and scrolls to it.
func (m *DiffModel) NextHunk() {
	if m.hunkCursor < len(m.hunks)-1 {
		m.hunkCursor++
		m.scrollToCurrentHunk()
	}
}

// PrevHunk moves the hunk cursor to the previous hunk and scrolls to it.
func (m *DiffModel) PrevHunk() {
	if m.hunkCursor > 0 {
		m.hunkCursor--
		m.scrollToCurrentHunk()
	}
}

// CurrentHunkIndex returns the index of the currently focused hunk.
func (m *DiffModel) CurrentHunkIndex() int {
	return m.hunkCursor
}

// CurrentHunk returns the currently focused hunk, or nil.
func (m *DiffModel) CurrentHunk() *git.DiffHunk {
	if m.hunkCursor < 0 || m.hunkCursor >= len(m.hunks) {
		return nil
	}
	return &m.hunks[m.hunkCursor]
}

// HunkCount returns the number of hunks.
func (m *DiffModel) HunkCount() int {
	return len(m.hunks)
}

// Hunks returns the raw hunk slice (for AI prompts etc.).
func (m *DiffModel) Hunks() []git.DiffHunk {
	return m.hunks
}

// BuildHunkPatch builds a minimal unified diff patch for a single hunk,
// suitable for piping to git apply --cached.
func (m *DiffModel) BuildHunkPatch(filePath string, hunkIdx int, reverse bool) string {
	if hunkIdx < 0 || hunkIdx >= len(m.hunks) {
		return ""
	}
	hunk := m.hunks[hunkIdx]

	var sb strings.Builder
	if !reverse {
		sb.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
		sb.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))
	} else {
		sb.WriteString(fmt.Sprintf("--- b/%s\n", filePath))
		sb.WriteString(fmt.Sprintf("+++ a/%s\n", filePath))
	}
	sb.WriteString(hunk.Header + "\n")

	for _, dl := range hunk.Lines {
		switch dl.Type {
		case git.DiffAdded:
			if reverse {
				sb.WriteString("-" + dl.Content + "\n")
			} else {
				sb.WriteString("+" + dl.Content + "\n")
			}
		case git.DiffRemoved:
			if reverse {
				sb.WriteString("+" + dl.Content + "\n")
			} else {
				sb.WriteString("-" + dl.Content + "\n")
			}
		case git.DiffContext:
			sb.WriteString(" " + dl.Content + "\n")
		}
	}
	return sb.String()
}

// View renders the diff panel content with a scrollbar when content overflows.
func (m *DiffModel) View() string {
	if len(m.lines) == 0 {
		return m.emptyStyle.Render("  Select a file to view diff")
	}

	viewH := m.Height
	if viewH <= 0 {
		viewH = 1
	}

	end := m.offset + viewH
	if end > len(m.lines) {
		end = len(m.lines)
	}
	visible := m.lines[m.offset:end]

	if len(m.lines) <= viewH {
		return strings.Join(visible, "\n")
	}

	// Render with scrollbar column on the right.
	scrollBar := renderDiffScrollbar(viewH, len(m.lines), m.offset)
	var sb strings.Builder
	for i, line := range visible {
		sb.WriteString(line)
		if i < len(scrollBar) {
			sb.WriteString(" ")
			sb.WriteString(scrollBar[i])
		}
		if i < len(visible)-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

func renderDiffScrollbar(height, total, offset int) []string {
	bars := make([]string, height)
	thumbSize := height * height / total
	if thumbSize < 1 {
		thumbSize = 1
	}
	maxOff := total - height
	thumbPos := 0
	if maxOff > 0 {
		thumbPos = offset * (height - thumbSize) / maxOff
	}
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a2a"))
	thumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#252538"))
	for i := range bars {
		if i >= thumbPos && i < thumbPos+thumbSize {
			bars[i] = thumbStyle.Render("▌")
		} else {
			bars[i] = trackStyle.Render(" ")
		}
	}
	return bars
}

// ScrollInfo returns "hunk X/Y" or "line X/Y" for the status bar.
func (m *DiffModel) ScrollInfo() string {
	if len(m.hunks) == 0 {
		return ""
	}
	return fmt.Sprintf("hunk %d/%d", m.hunkCursor+1, len(m.hunks))
}

func (m *DiffModel) buildLines() {
	m.lines = nil
	m.lineToHunk = nil
	m.hunkStarts = make([]int, len(m.hunks))

	w := m.Width - 4
	if w < 10 {
		w = 10
	}

	for hi, hunk := range m.hunks {
		m.hunkStarts[hi] = len(m.lines)
		isFocused := hi == m.hunkCursor

		for _, dl := range hunk.Lines {
			m.lines = append(m.lines, m.renderLine(dl, w, isFocused && dl.Type == git.DiffHunkHeader))
			m.lineToHunk = append(m.lineToHunk, hi)
		}
	}
}

func (m *DiffModel) renderLine(dl git.DiffLine, width int, focused bool) string {
	content := dl.Content
	if len(content) > width-2 {
		content = content[:width-3] + "…"
	}

	switch dl.Type {
	case git.DiffAdded:
		return m.addedStyle.Width(width).Render("+" + content)
	case git.DiffRemoved:
		return m.removedStyle.Width(width).Render("-" + content)
	case git.DiffHunkHeader:
		if focused {
			return m.hunkFocused.Width(width).Render("▶ " + content)
		}
		return m.hunkStyle.Render("  " + content)
	case git.DiffFileHeader:
		return m.headerStyle.Render(content)
	default:
		return m.contextStyle.Render(" " + content)
	}
}

func (m *DiffModel) scrollToCurrentHunk() {
	if m.hunkCursor >= len(m.hunkStarts) {
		return
	}
	target := m.hunkStarts[m.hunkCursor]
	// Rebuild so focused hunk header gets highlighted
	m.buildLines()
	// Scroll so hunk is near the top
	m.offset = target
	maxOffset := len(m.lines) - m.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}
