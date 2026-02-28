package panels

import (
	"fmt"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RemotesModel displays the list of configured git remotes.
type RemotesModel struct {
	remotes    []git.RemoteEntry
	list       widgets.ScrollList
	Width      int
	Height     int
	addModal   widgets.Modal
	showModal  bool
	addingURL  bool   // false = entering name, true = entering URL
	remoteName string // buffer while entering name
}

// NewRemotesModel creates a RemotesModel.
func NewRemotesModel(width, height int) RemotesModel {
	return RemotesModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-2, width-4),
	}
}

// SetRemotes updates the remotes list.
func (m *RemotesModel) SetRemotes(remotes []git.RemoteEntry) {
	m.remotes = remotes
	m.list.SetItems(m.buildItems())
}

// CurrentRemote returns the selected remote entry.
func (m *RemotesModel) CurrentRemote() *git.RemoteEntry {
	if len(m.remotes) == 0 || m.list.Cursor >= len(m.remotes) {
		return nil
	}
	return &m.remotes[m.list.Cursor]
}

// MoveUp moves the cursor up.
func (m *RemotesModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *RemotesModel) MoveDown() { m.list.MoveDown() }

// ShowAddModal opens the add-remote modal (name step).
func (m *RemotesModel) ShowAddModal() {
	m.addModal = widgets.NewInputModal("Add Remote — Name", "remote-name")
	m.addModal.Show()
	m.showModal = true
	m.addingURL = false
	m.remoteName = ""
}

// AdvanceToURL switches the modal to the URL input step.
func (m *RemotesModel) AdvanceToURL(name string) {
	m.remoteName = name
	m.addModal = widgets.NewInputModal("Add Remote — URL for "+name, "https://...")
	m.addModal.Show()
	m.addingURL = true
}

// IsModalVisible returns true if a modal is open.
func (m *RemotesModel) IsModalVisible() bool { return m.showModal }

// IsAddingURL returns true if the modal is in the URL-entry step.
func (m *RemotesModel) IsAddingURL() bool { return m.addingURL }

// RemoteName returns the buffered remote name.
func (m *RemotesModel) RemoteName() string { return m.remoteName }

// ModalInput returns the current modal text input value.
func (m *RemotesModel) ModalInput() string { return m.addModal.Input.Value() }

// HideModal hides the modal.
func (m *RemotesModel) HideModal() {
	m.addModal.Hide()
	m.showModal = false
	m.addingURL = false
	m.remoteName = ""
}

// UpdateModalInput forwards key events to the modal textinput.
func (m *RemotesModel) UpdateModalInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.addModal.Input, cmd = m.addModal.Input.Update(msg)
	return cmd
}

// SetModalInputModel replaces the modal's textinput (used by root Update).
func (m *RemotesModel) SetModalInputModel(ti textinput.Model) {
	m.addModal.Input = ti
}

// View renders the remotes panel.
func (m *RemotesModel) View() string {
	if len(m.remotes) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render("  No remotes configured")
	}
	return m.list.View()
}

// ModalView renders the active modal.
func (m *RemotesModel) ModalView() string { return m.addModal.View() }

func (m *RemotesModel) buildItems() []string {
	items := make([]string, len(m.remotes))
	for i, r := range m.remotes {
		items[i] = renderRemoteItem(r)
	}
	return items
}

func renderRemoteItem(r git.RemoteEntry) string {
	name := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true).Render(r.Name)

	url := r.FetchURL
	if url == "" {
		url = r.PushURL
	}
	urlStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Render(url)

	pushStr := ""
	if r.PushURL != "" && r.PushURL != r.FetchURL {
		pushStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).
			Render(fmt.Sprintf("  push: %s", r.PushURL))
	}

	return fmt.Sprintf(" ⬡ %s  %s%s", name, urlStr, pushStr)
}

// ListCursor returns the current scroll list cursor position.
func (m *RemotesModel) ListCursor() int { return m.list.Cursor }

// SetListCursor sets the scroll list cursor position.
func (m *RemotesModel) SetListCursor(n int) { m.list.Cursor = n }
