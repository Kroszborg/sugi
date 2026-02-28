package widgets

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// HelpSection is a named group of keybindings.
type HelpSection struct {
	Title    string
	Bindings []key.Binding
}

// HelpOverlay renders a scrollable help modal from sections.
type HelpOverlay struct {
	Sections     []HelpSection
	Width        int
	Height       int
	ScrollOffset int // lines scrolled down

	overlayStyle lipgloss.Style
	titleStyle   lipgloss.Style
	sectionStyle lipgloss.Style
	keyStyle     lipgloss.Style
	descStyle    lipgloss.Style
	footerStyle  lipgloss.Style
	scrollStyle  lipgloss.Style
}

// NewHelpOverlay creates a HelpOverlay for the given terminal size.
func NewHelpOverlay(width, height int) HelpOverlay {
	w := width - 4
	if w > 82 {
		w = 82
	}
	if w < 40 {
		w = 40
	}
	h := height - 4
	if h > 44 {
		h = 44
	}
	if h < 10 {
		h = 10
	}

	return HelpOverlay{
		Width:  w,
		Height: h,
		overlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4d9de0")).
			Padding(1, 2).
			Width(w).
			Height(h),
		titleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4d9de0")).
			Bold(true),
		sectionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e8835c")).
			Bold(true),
		keyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2ec4b6")).
			Bold(true).
			Width(18),
		descStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d8d8ee")),
		footerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3d3d5c")),
		scrollStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#252538")),
	}
}

func (h *HelpOverlay) ScrollDown() {
	h.ScrollOffset++
}

func (h *HelpOverlay) ScrollUp() {
	if h.ScrollOffset > 0 {
		h.ScrollOffset--
	}
}

// View renders the help overlay, truncating to fit the terminal.
func (h HelpOverlay) View() string {
	var sb strings.Builder

	sb.WriteString(h.titleStyle.Render("sugi — Keyboard Shortcuts") + "\n\n")

	for _, section := range h.Sections {
		divider := strings.Repeat("─", h.Width-len(section.Title)-5)
		sb.WriteString(h.sectionStyle.Render("  "+section.Title) + " " +
			h.scrollStyle.Render(divider) + "\n")
		for _, b := range section.Bindings {
			keyStr := strings.Join(b.Keys(), "/")
			sb.WriteString(
				h.keyStyle.Render("    "+keyStr) +
					h.descStyle.Render(b.Help().Desc) + "\n",
			)
		}
		sb.WriteString("\n")
	}

	allLines := strings.Split(sb.String(), "\n")
	totalContent := len(allLines)

	// Available lines inside the box: height minus border(2) and padding(2)
	avail := h.Height - 4
	if avail < 2 {
		avail = 2
	}
	// Reserve 1 line for footer
	contentLines := avail - 1

	// Apply scroll offset
	offset := h.ScrollOffset
	if offset < 0 {
		offset = 0
	}
	maxOffset := totalContent - contentLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	end := offset + contentLines
	if end > len(allLines) {
		end = len(allLines)
	}

	visible := allLines[offset:end]
	content := strings.Join(visible, "\n")

	// Build footer with scroll indicator
	footer := ""
	if totalContent > contentLines {
		pct := 0
		if maxOffset > 0 {
			pct = offset * 100 / maxOffset
		}
		scrollInfo := ""
		if offset > 0 {
			scrollInfo += "↑ "
		}
		scrollInfo += "scroll"
		if offset < maxOffset {
			scrollInfo += " ↓"
		}
		footer = h.footerStyle.Render("  [?/esc] close  [j/k] " + scrollInfo + " " + strings.Repeat("─", 4) + " " + itoa(pct) + "%")
	} else {
		footer = h.footerStyle.Render("  [?] or [esc] close help")
	}

	return h.overlayStyle.Render(content + "\n" + footer)
}
