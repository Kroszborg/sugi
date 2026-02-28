package widgets

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ModalKind distinguishes between a confirm dialog and an input dialog.
type ModalKind int

const (
	ModalConfirm ModalKind = iota
	ModalInput
)

// Modal is a simple centered overlay for confirmations and text input.
type Modal struct {
	Kind    ModalKind
	Title   string
	Body    string
	Input   textinput.Model
	Visible bool

	style      lipgloss.Style
	titleStyle lipgloss.Style
	bodyStyle  lipgloss.Style
}

// NewConfirmModal creates a yes/no confirmation modal.
func NewConfirmModal(title, body string) Modal {
	return Modal{
		Kind:  ModalConfirm,
		Title: title,
		Body:  body,
		style: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4d9de0")).
			Padding(1, 2).
			Width(50),
		titleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4d9de0")).
			Bold(true),
		bodyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d8d8ee")),
	}
}

// NewInputModal creates a text-input modal.
func NewInputModal(title, placeholder string) Modal {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 56

	return Modal{
		Kind:  ModalInput,
		Title: title,
		Input: ti,
		style: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4d9de0")).
			Padding(1, 2).
			Width(60),
		titleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4d9de0")).
			Bold(true),
		bodyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d8d8ee")),
	}
}

// View renders the modal. The result should be overlaid on the main view.
func (m Modal) View() string {
	if !m.Visible {
		return ""
	}

	title := m.titleStyle.Render(m.Title)

	var body string
	switch m.Kind {
	case ModalConfirm:
		confirmHint := lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")).Bold(true).Render("[y]") +
			m.bodyStyle.Render(" confirm") +
			"   " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).Render("[esc] cancel")
		body = m.bodyStyle.Render(m.Body) + "\n\n" + confirmHint
	case ModalInput:
		body = m.Input.View()
		if m.Body != "" {
			body += "\n\n" + m.bodyStyle.Render(m.Body)
		}
		inputHint := lipgloss.NewStyle().Foreground(lipgloss.Color("#7c6dfa")).Bold(true).Render("[enter]") +
			m.bodyStyle.Render(" confirm") +
			"   " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).Render("[esc] cancel")
		body += "\n\n" + inputHint
	}

	return m.style.Render(title + "\n\n" + body)
}

// Show makes the modal visible and (for input modals) focuses the input.
func (m *Modal) Show() {
	m.Visible = true
	if m.Kind == ModalInput {
		m.Input.Focus()
	}
}

// Hide hides the modal.
func (m *Modal) Hide() {
	m.Visible = false
	if m.Kind == ModalInput {
		m.Input.Reset()
		m.Input.Blur()
	}
}
