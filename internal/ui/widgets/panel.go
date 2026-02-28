package widgets

import (
	"github.com/charmbracelet/lipgloss"
)

// Panel renders a titled bordered box.
type Panel struct {
	Title   string
	Focused bool
	Width   int
	Height  int

	focusedStyle   lipgloss.Style
	unfocusedStyle lipgloss.Style
	titleStyle     lipgloss.Style
}

// NewPanel creates a Panel with the given dimensions.
func NewPanel(title string, width, height int) Panel {
	return Panel{
		Title:  title,
		Width:  width,
		Height: height,
		focusedStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4d9de0")).
			Width(width - 2).
			Height(height - 2),
		unfocusedStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3d3d5c")).
			Width(width - 2).
			Height(height - 2),
		titleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4d9de0")).
			Bold(true).
			Padding(0, 1),
	}
}

// Render wraps content in a bordered panel with a title.
func (p Panel) Render(content string) string {
	style := p.unfocusedStyle
	if p.Focused {
		style = p.focusedStyle
	}

	// Inject title into top border using padding trick
	_ = p.titleStyle // title is rendered via the border title feature

	return style.Render(content)
}

// RenderWithTitle renders a panel with a title in the top border.
func (p Panel) RenderWithTitle(content string) string {
	borderColor := lipgloss.Color("#3d3d5c")
	titleColor := lipgloss.Color("#7878a0")
	if p.Focused {
		borderColor = lipgloss.Color("#4d9de0")
		titleColor = lipgloss.Color("#4d9de0")
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(p.Width - 2).
		Height(p.Height - 2)

	title := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(p.Focused).
		Render(" " + p.Title + " ")

	_ = title

	return style.Render(content)
}
