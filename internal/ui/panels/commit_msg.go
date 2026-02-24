package panels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// CommitMsgModel handles the commit message editing panel.
type CommitMsgModel struct {
	Textarea textarea.Model // exported for root model Update
	Width    int
	Height   int
	Focused  bool

	borderStyle lipgloss.Style
	labelStyle  lipgloss.Style
	hintStyle   lipgloss.Style
}

// NewCommitMsgModel creates a commit message panel.
func NewCommitMsgModel(width, height int) CommitMsgModel {
	ta := textarea.New()
	ta.Placeholder = "Commit message (conventional commits: feat: add login)"
	ta.SetWidth(width - 6)
	ta.SetHeight(height - 6)
	ta.CharLimit = 5000
	ta.ShowLineNumbers = false

	return CommitMsgModel{
		Width:    width,
		Height:   height,
		Textarea: ta,
		borderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Padding(0, 1),
		labelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true),
		hintStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70")),
	}
}

// Focus focuses the textarea for editing.
func (m *CommitMsgModel) Focus() {
	m.Focused = true
	m.Textarea.Focus()
}

// Blur unfocuses the textarea.
func (m *CommitMsgModel) Blur() {
	m.Focused = false
	m.Textarea.Blur()
}

// Value returns the current commit message text.
func (m *CommitMsgModel) Value() string {
	return m.Textarea.Value()
}

// SetValue sets the commit message text (e.g. from AI generation).
func (m *CommitMsgModel) SetValue(s string) {
	m.Textarea.SetValue(s)
}

// AppendText appends text to the commit message (for streaming AI output).
func (m *CommitMsgModel) AppendText(s string) {
	current := m.Textarea.Value()
	m.Textarea.SetValue(current + s)
}

// Reset clears the commit message.
func (m *CommitMsgModel) Reset() {
	m.Textarea.Reset()
}

// Update forwards a Bubbletea message to the textarea and returns any cmd.
func (m *CommitMsgModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.Textarea, cmd = m.Textarea.Update(msg)
	return cmd
}

// View renders the commit message panel.
func (m *CommitMsgModel) View() string {
	label := m.labelStyle.Render("Commit Message")
	hint := m.hintStyle.Render("  [ctrl+s] commit  [esc] cancel")

	return m.borderStyle.
		Width(m.Width - 4).
		Height(m.Height - 4).
		Render(label + "\n\n" + m.Textarea.View() + "\n" + hint)
}
