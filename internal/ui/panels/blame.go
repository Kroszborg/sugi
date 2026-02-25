package panels

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/charmbracelet/lipgloss"
)

// BlameModel displays git blame output for a file.
type BlameModel struct {
	lines  []git.BlameLine
	items  []string // rendered lines
	offset int
	Width  int
	Height int

	cursorStyle  lipgloss.Style
	normalStyle  lipgloss.Style
	hashStyle    lipgloss.Style
	authorStyle  lipgloss.Style
	dateStyle    lipgloss.Style
	contentStyle lipgloss.Style
	dimStyle     lipgloss.Style
	emptyStyle   lipgloss.Style

	cursor int
}

// NewBlameModel creates a BlameModel.
func NewBlameModel(width, height int) BlameModel {
	return BlameModel{
		Width:        width,
		Height:       height,
		cursorStyle:  lipgloss.NewStyle().Background(lipgloss.Color("#313244")).Foreground(lipgloss.Color("#cdd6f4")),
		normalStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
		hashStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true),
		authorStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")),
		dateStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")),
		contentStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
		dimStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a")),
		emptyStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")),
	}
}

// SetBlame updates the blame lines.
func (m *BlameModel) SetBlame(lines []git.BlameLine) {
	m.lines = lines
	m.items = m.buildItems()
	m.cursor = 0
	m.offset = 0
}

// CurrentLine returns the blame line at the cursor.
func (m *BlameModel) CurrentLine() *git.BlameLine {
	if len(m.lines) == 0 || m.cursor >= len(m.lines) {
		return nil
	}
	return &m.lines[m.cursor]
}

// MoveUp moves the cursor up.
func (m *BlameModel) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
		m.clampOffset()
	}
}

// MoveDown moves the cursor down.
func (m *BlameModel) MoveDown() {
	if m.cursor < len(m.lines)-1 {
		m.cursor++
		m.clampOffset()
	}
}

// PageUp scrolls up a page.
func (m *BlameModel) PageUp() {
	m.cursor -= m.Height
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.clampOffset()
}

// PageDown scrolls down a page.
func (m *BlameModel) PageDown() {
	m.cursor += m.Height
	if m.cursor >= len(m.lines) {
		m.cursor = len(m.lines) - 1
	}
	m.clampOffset()
}

// View renders the blame panel.
func (m *BlameModel) View() string {
	if len(m.items) == 0 {
		return m.emptyStyle.Render("  No blame data. Select a file and press b.")
	}

	end := m.offset + m.Height
	if end > len(m.items) {
		end = len(m.items)
	}

	var sb strings.Builder
	for i := m.offset; i < end; i++ {
		line := m.items[i]
		if i == m.cursor {
			// Highlight cursor row with background
			sb.WriteString(m.cursorStyle.Width(m.Width - 4).Render(line))
		} else {
			sb.WriteString(line)
		}
		if i < end-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// StatusInfo returns "line X/Y" for the status bar.
func (m *BlameModel) StatusInfo() string {
	if len(m.lines) == 0 {
		return ""
	}
	return fmt.Sprintf("line %d/%d", m.cursor+1, len(m.lines))
}

func (m *BlameModel) buildItems() []string {
	items := make([]string, len(m.lines))
	// Max author length for alignment
	maxAuthor := 12
	for _, l := range m.lines {
		name := l.Author
		if len(name) > 16 {
			name = name[:16]
		}
		if len(name) > maxAuthor {
			maxAuthor = len(name)
		}
	}
	if maxAuthor > 16 {
		maxAuthor = 16
	}

	prevHash := ""
	for i, l := range m.lines {
		hash := ""
		author := ""
		date := ""

		if l.Hash != prevHash {
			hash = m.hashStyle.Render(l.ShortHash)
			a := l.Author
			if len(a) > maxAuthor {
				a = a[:maxAuthor-1] + "…"
			}
			author = m.authorStyle.Render(fmt.Sprintf("%-*s", maxAuthor, a))
			if !l.Date.IsZero() {
				date = m.dateStyle.Render(l.Date.Format("2006-01-02"))
			}
			prevHash = l.Hash
		} else {
			hash = m.dimStyle.Render(strings.Repeat(" ", 8))
			author = m.dimStyle.Render(strings.Repeat(" ", maxAuthor))
			date = m.dimStyle.Render("          ")
		}

		lineNum := m.dateStyle.Render(fmt.Sprintf("%4d", l.LineNum))
		content := l.Content
		// Truncate content to fit
		maxContent := m.Width - 8 - maxAuthor - 10 - 4 - 6
		if maxContent > 10 && len(content) > maxContent {
			content = content[:maxContent-1] + "…"
		}

		items[i] = fmt.Sprintf("%s %s %s %s  %s",
			hash, author, date, lineNum, m.contentStyle.Render(content))
	}
	return items
}

func (m *BlameModel) clampOffset() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.Height {
		m.offset = m.cursor - m.Height + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}
