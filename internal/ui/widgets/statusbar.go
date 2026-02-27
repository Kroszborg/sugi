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
	Width    int
	Hints    []KeyHint
	Extra    string // right-aligned status message
	ModePill string // optional colored mode badge (e.g. "REBASE")

	keyStyle   lipgloss.Style
	descStyle  lipgloss.Style
	sepStyle   lipgloss.Style
	dotStyle   lipgloss.Style
	barStyle   lipgloss.Style
	extraStyle lipgloss.Style
}

// NewStatusBar creates a StatusBar for the given terminal width.
func NewStatusBar(width int) StatusBar {
	bg := lipgloss.Color("#181825")
	return StatusBar{
		Width: width,
		keyStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#89dceb")).
			Bold(true),
		descStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#6c7086")),
		sepStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#313244")),
		dotStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#45475a")),
		barStyle: lipgloss.NewStyle().
			Background(bg).
			Width(width),
		extraStyle: lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color("#6c7086")),
	}
}

// View renders the status bar.
func (sb StatusBar) View() string {
	dot := sb.dotStyle.Render(" · ")

	var parts []string
	for _, h := range sb.Hints {
		parts = append(parts,
			sb.keyStyle.Render(h.Key)+
				sb.sepStyle.Render(":")+
				sb.descStyle.Render(h.Desc),
		)
	}
	left := strings.Join(parts, dot)

	// Mode pill (e.g. "MERGE" "REBASE") shown on the right before Extra
	pill := ""
	if sb.ModePill != "" {
		pill = sb.ModePill + "  "
	}

	right := pill + sb.extraStyle.Render(sb.Extra)

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	pad := sb.Width - leftLen - rightLen - 1
	if pad < 1 {
		pad = 1
	}

	return sb.barStyle.Render(left + strings.Repeat(" ", pad) + right)
}

// ModePillStyle returns a colored badge for the given mode name.
func ModePillStyle(label, fg, bg string) string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(fg)).
		Bold(true).
		Padding(0, 1).
		Render(label)
}
