package widgets

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// KeyHint is a single key+description pair shown in the status bar.
type KeyHint struct {
	Key  string
	Desc string
}

// StatusBar renders the bottom key-hint bar.
type StatusBar struct {
	Width int
	Hints []KeyHint
	Extra string // right-aligned status message

	keyStyle   lipgloss.Style
	descStyle  lipgloss.Style
	sepStyle   lipgloss.Style
	barStyle   lipgloss.Style
	extraStyle lipgloss.Style
}

// NewStatusBar creates a StatusBar for the given terminal width.
func NewStatusBar(width int) StatusBar {
	bg := lipgloss.Color("#181825") // ColorMantle — matches header
	return StatusBar{
		Width: width,
		keyStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#89dceb")). // ColorSky
			Bold(true),
		descStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#585b70")), // ColorMuted
		sepStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#313244")), // ColorSurface
		barStyle: lipgloss.NewStyle().
			Background(bg).
			Width(width),
		extraStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#585b70")),
	}
}

// View renders the status bar.
func (sb StatusBar) View() string {
	var parts []string
	for _, h := range sb.Hints {
		parts = append(parts,
			sb.keyStyle.Render(h.Key)+
				sb.sepStyle.Render(":")+
				sb.descStyle.Render(h.Desc),
		)
	}
	left := strings.Join(parts, sb.sepStyle.Render("  "))
	right := sb.extraStyle.Render(sb.Extra)

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	pad := sb.Width - leftLen - rightLen
	if pad < 1 {
		pad = 1
	}

	return sb.barStyle.Render(left + strings.Repeat(" ", pad) + right)
}
