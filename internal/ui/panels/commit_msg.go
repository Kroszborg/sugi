package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommitField tracks which field is active in the commit form.
type CommitField int

const (
	CommitFieldSubject CommitField = iota
	CommitFieldBody
)

// CommitMsgModel is a two-field commit form: subject line + body.
type CommitMsgModel struct {
	subject      textinput.Model
	body         textarea.Model
	active       CommitField
	Width        int
	Height       int
	Focused      bool
	AIGenerating bool // true while AI is streaming
}

// NewCommitMsgModel creates a two-field commit form.
func NewCommitMsgModel(width, height int) CommitMsgModel {
	innerW := width - 8
	if innerW < 20 {
		innerW = 20
	}

	subj := textinput.New()
	subj.Placeholder = "feat(scope): brief description"
	subj.CharLimit = 72
	subj.Width = innerW - 2

	subj.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#89dceb"))
	subj.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	subj.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a"))

	bodyH := height - 14
	if bodyH < 3 {
		bodyH = 3
	}

	body := textarea.New()
	body.Placeholder = "Explain why this change was made…"
	body.SetWidth(innerW - 2)
	body.SetHeight(bodyH)
	body.CharLimit = 5000
	body.ShowLineNumbers = false
	body.FocusedStyle.Base = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	body.BlurredStyle.Base = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8"))

	return CommitMsgModel{
		Width:   width,
		Height:  height,
		subject: subj,
		body:    body,
		active:  CommitFieldSubject,
	}
}

// Focus activates the commit form, starting at the subject field.
func (m *CommitMsgModel) Focus() {
	m.Focused = true
	m.active = CommitFieldSubject
	m.subject.Focus()
	m.body.Blur()
}

// Blur deactivates the commit form.
func (m *CommitMsgModel) Blur() {
	m.Focused = false
	m.subject.Blur()
	m.body.Blur()
}

// NextField cycles to the next input field (Tab).
func (m *CommitMsgModel) NextField() {
	if m.active == CommitFieldSubject {
		m.active = CommitFieldBody
		m.subject.Blur()
		m.body.Focus()
	} else {
		m.active = CommitFieldSubject
		m.body.Blur()
		m.subject.Focus()
	}
}

// Value returns the full commit message: subject + blank line + body (if any).
func (m *CommitMsgModel) Value() string {
	subj := strings.TrimSpace(m.subject.Value())
	body := strings.TrimSpace(m.body.Value())
	if subj == "" {
		return ""
	}
	if body == "" {
		return subj
	}
	return subj + "\n\n" + body
}

// Subject returns just the subject line.
func (m *CommitMsgModel) Subject() string {
	return strings.TrimSpace(m.subject.Value())
}

// SetValue populates the form from a full commit message string.
func (m *CommitMsgModel) SetValue(s string) {
	parts := strings.SplitN(s, "\n\n", 2)
	subj := strings.TrimSpace(parts[0])
	m.subject.SetValue(subj)
	if len(parts) > 1 {
		m.body.SetValue(strings.TrimSpace(parts[1]))
	}
}

// AppendToSubject appends text to the subject (for AI streaming).
func (m *CommitMsgModel) AppendToSubject(s string) {
	m.subject.SetValue(m.subject.Value() + s)
}

// Reset clears both fields.
func (m *CommitMsgModel) Reset() {
	m.subject.Reset()
	m.body.Reset()
	m.active = CommitFieldSubject
	m.AIGenerating = false
}

// Update forwards a message to whichever field is active.
func (m *CommitMsgModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.active {
	case CommitFieldSubject:
		m.subject, cmd = m.subject.Update(msg)
	case CommitFieldBody:
		m.body, cmd = m.body.Update(msg)
	}
	return cmd
}

// View renders the commit form.
func (m *CommitMsgModel) View() string {
	innerW := m.Width - 8
	if innerW < 20 {
		innerW = 20
	}

	// ── Colors ───────────────────────────────────────────────
	clrSky     := lipgloss.Color("#89dceb")
	clrTeal    := lipgloss.Color("#94e2d5")
	clrSurface := lipgloss.Color("#313244")
	clrOverlay := lipgloss.Color("#45475a")
	clrMuted   := lipgloss.Color("#585b70")
	clrText    := lipgloss.Color("#cdd6f4")
	clrYellow  := lipgloss.Color("#f9e2af")
	clrRed     := lipgloss.Color("#f38ba8")
	clrGreen   := lipgloss.Color("#a6e3a1")

	// ── Subject ──────────────────────────────────────────────
	charCount := len(m.subject.Value())

	var counterFg lipgloss.Color
	var counterText string
	switch {
	case charCount > 72:
		counterFg = clrRed
		counterText = fmt.Sprintf("%d/72 !", charCount)
	case charCount > 50:
		counterFg = clrYellow
		counterText = fmt.Sprintf("%d/72", charCount)
	default:
		counterFg = clrOverlay
		counterText = fmt.Sprintf("%d/72", charCount)
	}

	subjectActive := m.active == CommitFieldSubject
	var subjBorderClr lipgloss.Color
	var subjLabelClr lipgloss.Color
	if subjectActive && !m.AIGenerating {
		subjBorderClr = clrSky
		subjLabelClr = clrSky
	} else if m.AIGenerating {
		subjBorderClr = clrTeal
		subjLabelClr = clrTeal
	} else {
		subjBorderClr = clrSurface
		subjLabelClr = clrMuted
	}

	// Label row: icon + SUBJECT + spacer + counter
	labelLeft := lipgloss.NewStyle().Foreground(subjLabelClr).Bold(subjectActive || m.AIGenerating).
		Render("◆ SUBJECT")
	counterRendered := lipgloss.NewStyle().Foreground(counterFg).Bold(charCount > 72).
		Render(counterText)
	spacerW := innerW - lipgloss.Width(labelLeft) - lipgloss.Width(counterRendered) - 2
	if spacerW < 1 {
		spacerW = 1
	}
	spacer := strings.Repeat(" ", spacerW)
	subjLabelRow := labelLeft + spacer + counterRendered

	subjBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(subjBorderClr).
		Padding(0, 1).
		Width(innerW).
		Render(m.subject.View())

	// ── Body ─────────────────────────────────────────────────
	bodyActive := m.active == CommitFieldBody
	var bodyBorderClr lipgloss.Color
	var bodyLabelClr lipgloss.Color
	if m.AIGenerating {
		bodyBorderClr = clrTeal
		bodyLabelClr = clrTeal
	} else if bodyActive {
		bodyBorderClr = clrSky
		bodyLabelClr = clrSky
	} else {
		bodyBorderClr = clrSurface
		bodyLabelClr = clrMuted
	}

	bodyLabelMain := lipgloss.NewStyle().Foreground(bodyLabelClr).Bold(bodyActive || m.AIGenerating).
		Render("◆ DESCRIPTION")

	var bodyLabelSuffix string
	if m.AIGenerating {
		bodyLabelSuffix = lipgloss.NewStyle().Foreground(clrTeal).Render("  ✦ writing…")
	} else {
		bodyLabelSuffix = lipgloss.NewStyle().Foreground(clrOverlay).Render("  explain why, not what")
	}
	bodyLabelRow := bodyLabelMain + bodyLabelSuffix

	bodyBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(bodyBorderClr).
		Padding(0, 1).
		Width(innerW).
		Render(m.body.View())

	// ── Commit type guide ─────────────────────────────────────
	typeGuide := ""
	if charCount == 0 && !m.AIGenerating {
		types := []struct{ t, clr string }{
			{"feat", "#a6e3a1"}, {"fix", "#f38ba8"}, {"docs", "#89b4fa"},
			{"refactor", "#cba6f7"}, {"chore", "#585b70"},
		}
		var parts []string
		for _, ty := range types {
			parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color(ty.clr)).Render(ty.t))
		}
		typeGuide = lipgloss.NewStyle().Foreground(clrOverlay).Render("  ") +
			strings.Join(parts, lipgloss.NewStyle().Foreground(clrOverlay).Render(" · "))
	}

	// ── Char count status bar ─────────────────────────────────
	var charBar string
	if charCount > 0 {
		filled := charCount * (innerW - 4) / 72
		if filled > innerW-4 {
			filled = innerW - 4
		}
		empty := innerW - 4 - filled
		barClr := clrGreen
		if charCount > 72 {
			barClr = clrRed
		} else if charCount > 50 {
			barClr = clrYellow
		}
		charBar = "  " +
			lipgloss.NewStyle().Foreground(barClr).Render(strings.Repeat("▪", filled)) +
			lipgloss.NewStyle().Foreground(clrSurface).Render(strings.Repeat("▪", empty))
	}

	// ── Hints ─────────────────────────────────────────────────
	dimStyle := lipgloss.NewStyle().Foreground(clrOverlay)
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Bold(true)

	sep := dimStyle.Render("  ·  ")
	hints := "  " +
		keyStyle.Render("tab") + dimStyle.Render(" switch") +
		sep +
		keyStyle.Render("ctrl+g") + dimStyle.Render("/") + keyStyle.Render("alt+g") + dimStyle.Render(" AI") +
		sep +
		keyStyle.Render("ctrl+s") + dimStyle.Render(" commit") +
		sep +
		keyStyle.Render("esc") + dimStyle.Render(" cancel")

	// ── Divider ───────────────────────────────────────────────
	divStyle := lipgloss.NewStyle().Foreground(clrSurface)
	div := divStyle.Render(strings.Repeat("─", innerW+2))

	rows := []string{
		subjLabelRow,
		subjBox,
	}
	if charBar != "" {
		rows = append(rows, charBar)
	}
	if typeGuide != "" {
		rows = append(rows, typeGuide)
	}
	rows = append(rows,
		"",
		bodyLabelRow,
		bodyBox,
		"",
		div,
		hints,
	)

	// Filter out leading/trailing empty
	_ = clrText
	return strings.Join(rows, "\n")
}
