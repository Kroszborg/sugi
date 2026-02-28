package panels

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BranchModel displays the list of local branches.
type BranchModel struct {
	branches []git.Branch
	list     widgets.ScrollList
	Width    int
	Height   int

	newBranchModal widgets.Modal
	showingModal   bool
}

// NewBranchModel creates a BranchModel with the given dimensions.
func NewBranchModel(width, height int) BranchModel {
	return BranchModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-3, width-4),
	}
}

// SetBranches updates the branch list.
func (m *BranchModel) SetBranches(branches []git.Branch) {
	m.branches = branches
	m.list.SetItems(m.buildItems())
}

// CurrentBranch returns the Branch at the cursor, or nil.
func (m *BranchModel) CurrentBranch() *git.Branch {
	if len(m.branches) == 0 || m.list.Cursor >= len(m.branches) {
		return nil
	}
	return &m.branches[m.list.Cursor]
}

// MoveUp moves the cursor up.
func (m *BranchModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *BranchModel) MoveDown() { m.list.MoveDown() }

// ShowNewBranchModal opens the new-branch input modal.
func (m *BranchModel) ShowNewBranchModal() {
	m.newBranchModal = widgets.NewInputModal("New Branch", "branch-name")
	m.newBranchModal.Show()
	m.showingModal = true
}

// ShowRenameBranchModal opens the rename-branch modal pre-filled with currentName.
func (m *BranchModel) ShowRenameBranchModal(currentName string) {
	m.newBranchModal = widgets.NewInputModal("Rename Branch", currentName)
	m.newBranchModal.Input.SetValue(currentName)
	m.newBranchModal.Show()
	m.showingModal = true
}

// HideModal hides any open modal.
func (m *BranchModel) HideModal() {
	m.newBranchModal.Hide()
	m.showingModal = false
}

// IsModalVisible returns true if a modal is open.
func (m *BranchModel) IsModalVisible() bool {
	return m.showingModal
}

// ModalInput returns the current text in the new-branch input.
func (m *BranchModel) ModalInput() string {
	return m.newBranchModal.Input.Value()
}

// UpdateModalInput forwards a message to the modal's textinput and returns any cmd.
func (m *BranchModel) UpdateModalInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.newBranchModal.Input, cmd = m.newBranchModal.Input.Update(msg)
	return cmd
}

// SetModalInputModel replaces the modal input (used by root Update).
func (m *BranchModel) SetModalInputModel(ti textinput.Model) {
	m.newBranchModal.Input = ti
}

// View renders the branches panel content.
func (m *BranchModel) View() string {
	if len(m.branches) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).Render("  No branches")
	}
	return m.list.View()
}

// ModalView renders the active modal if visible.
func (m *BranchModel) ModalView() string {
	return m.newBranchModal.View()
}

func (m *BranchModel) buildItems() []string {
	items := make([]string, len(m.branches))
	for i, b := range m.branches {
		items[i] = renderBranchItem(b)
	}
	return items
}

func renderBranchItem(b git.Branch) string {
	marker := widgets.BranchBadge(b.IsCurrent)
	name := b.Name
	if b.IsCurrent {
		name = lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")).Bold(true).Render(name)
	} else {
		name = lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee")).Render(name)
	}

	tracking := ""
	if b.Upstream != "" {
		ab := widgets.AheadBehindBadge(b.Ahead, b.Behind)
		if ab != "" {
			tracking = " " + ab
		} else {
			tracking = lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).Render(" ✓")
		}
	}

	upstream := ""
	if b.Upstream != "" {
		short := b.Upstream
		if idx := strings.Index(short, "/"); idx >= 0 {
			short = short[idx+1:]
		}
		upstream = lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c")).Render(fmt.Sprintf(" [%s]", short))
	}

	return fmt.Sprintf(" %s %s%s%s", marker, name, tracking, upstream)
}

// ListCursor returns the current scroll list cursor position.
func (m *BranchModel) ListCursor() int { return m.list.Cursor }

// SetListCursor sets the scroll list cursor position.
func (m *BranchModel) SetListCursor(n int) { m.list.Cursor = n }

// GetBranches returns the raw branch slice (for rebuildPanels preservation).
func (m *BranchModel) GetBranches() []git.Branch { return m.branches }
