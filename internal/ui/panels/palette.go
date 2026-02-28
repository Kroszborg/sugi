package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaletteSelectMsg is emitted when the user confirms a palette entry.
type PaletteSelectMsg struct{ ID string }

// PaletteEntry is a single command in the palette.
type PaletteEntry struct {
	ID       string
	Label    string
	Keys     string
	Category string
}

// PaletteModel is the fuzzy command palette overlay.
type PaletteModel struct {
	all      []PaletteEntry
	filtered []PaletteEntry
	cursor   int
	input    textinput.Model
	Width    int
	Height   int

	titleStyle    lipgloss.Style
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	keysStyle     lipgloss.Style
	catStyle      lipgloss.Style
	dimStyle      lipgloss.Style
	boxStyle      lipgloss.Style
}

// NewPaletteModel creates a PaletteModel.
func NewPaletteModel(width, height int) PaletteModel {
	ti := textinput.New()
	ti.Placeholder = "Search commands…"
	ti.CharLimit = 100
	ti.Width = width - 12

	return PaletteModel{
		Width:  width,
		Height: height,
		input:  ti,

		titleStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#4d9de0")).Bold(true),
		selectedStyle: lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2a")).Foreground(lipgloss.Color("#d8d8ee")),
		normalStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee")),
		keysStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")),
		catStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")),
		dimStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#252538")),
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4d9de0")).
			Padding(0, 1),
	}
}

// SetEntries sets the full list of palette entries.
func (m *PaletteModel) SetEntries(entries []PaletteEntry) {
	m.all = entries
	m.applyFilter()
}

// Open resets the palette and focuses the input.
func (m *PaletteModel) Open() {
	m.input.Reset()
	m.input.Focus()
	m.cursor = 0
	m.applyFilter()
}

// MoveUp moves the cursor up.
func (m *PaletteModel) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveDown moves the cursor down.
func (m *PaletteModel) MoveDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
	}
}

// Current returns the currently highlighted entry, or nil.
func (m *PaletteModel) Current() *PaletteEntry {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return &m.filtered[m.cursor]
}

// Update forwards key input to the text field and re-filters.
func (m *PaletteModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.cursor = 0
	m.applyFilter()
	return cmd
}

func (m *PaletteModel) applyFilter() {
	q := strings.ToLower(m.input.Value())
	if q == "" {
		m.filtered = make([]PaletteEntry, len(m.all))
		copy(m.filtered, m.all)
		return
	}
	m.filtered = m.filtered[:0]
	for _, e := range m.all {
		if strings.Contains(strings.ToLower(e.Label), q) ||
			strings.Contains(strings.ToLower(e.Category), q) ||
			strings.Contains(strings.ToLower(e.Keys), q) {
			m.filtered = append(m.filtered, e)
		}
	}
}

// View renders the palette overlay box.
func (m *PaletteModel) View() string {
	w := m.Width - 6
	if w < 40 {
		w = 40
	}
	maxVisible := m.Height - 9
	if maxVisible < 3 {
		maxVisible = 3
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	labelW := w - 26
	if labelW < 10 {
		labelW = 10
	}

	var sb strings.Builder
	sb.WriteString(m.titleStyle.Render("⌘ Command Palette") + "\n")
	sb.WriteString(m.input.View() + "\n")
	sb.WriteString(m.dimStyle.Render(strings.Repeat("─", w)) + "\n")

	if len(m.filtered) == 0 {
		sb.WriteString(m.dimStyle.Render("  no matching commands"))
	} else {
		for i := start; i < end; i++ {
			e := m.filtered[i]
			label := e.Label
			if len(label) > labelW {
				label = label[:labelW-1] + "…"
			}
			keys := m.keysStyle.Render(fmt.Sprintf("%-12s", e.Keys))
			cat := m.catStyle.Render(e.Category)
			line := fmt.Sprintf("  %-*s  %s  %s", labelW, label, keys, cat)
			if i == m.cursor {
				sb.WriteString(m.selectedStyle.Width(w).Render(line))
			} else {
				sb.WriteString(m.normalStyle.Render(line))
			}
			sb.WriteString("\n")
		}
		counter := fmt.Sprintf("  %d/%d commands", len(m.filtered), len(m.all))
		sb.WriteString(m.dimStyle.Render(counter))
	}

	return m.boxStyle.Width(w).Render(sb.String())
}
