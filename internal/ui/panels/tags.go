package panels

import (
	"fmt"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TagsModel displays the list of local tags.
type TagsModel struct {
	tags   []git.Tag
	list   widgets.ScrollList
	Width  int
	Height int

	newTagModal  widgets.Modal
	showingModal bool
}

// NewTagsModel creates a TagsModel.
func NewTagsModel(width, height int) TagsModel {
	return TagsModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-2, width-4),
	}
}

// SetTags updates the tag list.
func (m *TagsModel) SetTags(tags []git.Tag) {
	m.tags = tags
	m.list.SetItems(m.buildItems())
}

// CurrentTag returns the tag at the cursor, or nil.
func (m *TagsModel) CurrentTag() *git.Tag {
	if len(m.tags) == 0 || m.list.Cursor >= len(m.tags) {
		return nil
	}
	return &m.tags[m.list.Cursor]
}

// MoveUp moves the cursor up.
func (m *TagsModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *TagsModel) MoveDown() { m.list.MoveDown() }

// ShowNewTagModal opens the new-tag input modal.
func (m *TagsModel) ShowNewTagModal() {
	m.newTagModal = widgets.NewInputModal("New Tag", "v1.0.0")
	m.newTagModal.Show()
	m.showingModal = true
}

// HideModal hides any open modal.
func (m *TagsModel) HideModal() {
	m.newTagModal.Hide()
	m.showingModal = false
}

// IsModalVisible returns true if the new-tag modal is open.
func (m *TagsModel) IsModalVisible() bool {
	return m.showingModal
}

// ModalInput returns the current text in the new-tag input.
func (m *TagsModel) ModalInput() string {
	return m.newTagModal.Input.Value()
}

// UpdateModalInput forwards a key message to the modal's textinput.
func (m *TagsModel) UpdateModalInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.newTagModal.Input, cmd = m.newTagModal.Input.Update(msg)
	return cmd
}

// View renders the tags panel.
func (m *TagsModel) View() string {
	if len(m.tags) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render("  No tags")
	}
	return m.list.View()
}

// ModalView renders the active modal (if visible).
func (m *TagsModel) ModalView() string {
	return m.newTagModal.View()
}

func (m *TagsModel) buildItems() []string {
	items := make([]string, len(m.tags))
	for i, t := range m.tags {
		hash := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Render(t.ShortHash)

		date := ""
		if !t.Date.IsZero() {
			date = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(relativeDate(t.Date))
		}

		kind := ""
		if t.IsAnnotated {
			kind = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render(" ◆")
		} else {
			kind = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(" ○")
		}

		name := lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Bold(true).Render(t.Name)

		msg := ""
		if t.Message != "" {
			msg = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")).Render("  " + t.Message)
		}

		items[i] = fmt.Sprintf(" %s %s %s%s%s", kind, name, hash, date, msg)
	}
	return items
}
