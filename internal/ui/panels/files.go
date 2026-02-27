package panels

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// FilesModel displays staged/unstaged/untracked files with status badges.
type FilesModel struct {
	files  []git.FileStatus
	list   widgets.ScrollList
	Width  int
	Height int

	headerStyle lipgloss.Style
	sectionStyle lipgloss.Style
}

// NewFilesModel creates a FilesModel with the given dimensions.
func NewFilesModel(width, height int) FilesModel {
	return FilesModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-2, width-4),
		headerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70")).
			Bold(true),
		sectionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70")),
	}
}

// SetFiles updates the file list.
func (m *FilesModel) SetFiles(files []git.FileStatus) {
	m.files = files
	m.list.SetItems(m.buildItems())
}

// CurrentFile returns the FileStatus at the cursor, or nil.
func (m *FilesModel) CurrentFile() *git.FileStatus {
	if len(m.files) == 0 {
		return nil
	}
	// Map list cursor back to files index (items include section headers)
	itemIdx := m.list.Cursor
	fileIdx := 0
	for i, item := range m.list.Items {
		if i == itemIdx {
			break
		}
		if !isSectionHeader(item) {
			fileIdx++
		}
	}
	// Walk files to find fileIdx-th non-header item
	count := -1
	for i, item := range m.list.Items {
		if isSectionHeader(item) {
			continue
		}
		count++
		if i == itemIdx {
			if count < len(m.files) {
				return &m.files[count]
			}
		}
	}
	return nil
}

// MoveUp moves the cursor up.
func (m *FilesModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the cursor down.
func (m *FilesModel) MoveDown() { m.list.MoveDown() }

// ToggleSelect toggles multi-selection on the current file item.
func (m *FilesModel) ToggleSelect() {
	// Only allow selecting non-header items
	cur := m.list.CurrentItem()
	if isSectionHeader(cur) {
		return
	}
	m.list.ToggleSelect()
}

// HasSelection returns true if any files are multi-selected.
func (m *FilesModel) HasSelection() bool { return m.list.HasSelection() }

// SelectionCount returns the number of multi-selected items.
func (m *FilesModel) SelectionCount() int { return len(m.list.SelectedIndices()) }

// SelectedFiles returns the FileStatus entries for all multi-selected list rows.
func (m *FilesModel) SelectedFiles() []git.FileStatus {
	indices := m.list.SelectedIndices()
	var out []git.FileStatus
	fileIdx := 0
	for i, item := range m.list.Items {
		if isSectionHeader(item) {
			continue
		}
		for _, sel := range indices {
			if sel == i && fileIdx < len(m.files) {
				out = append(out, m.files[fileIdx])
			}
		}
		fileIdx++
	}
	return out
}

// ClearSelection clears multi-selection.
func (m *FilesModel) ClearSelection() { m.list.ClearSelected() }

// View renders the files panel content (without border).
func (m *FilesModel) View() string {
	if len(m.files) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70")).
			Render("  No changes")
	}
	return m.list.View()
}

// Stats returns a summary string like "3 staged  2 unstaged  1 untracked".
func (m *FilesModel) Stats() string {
	staged, unstaged, untracked := 0, 0, 0
	for _, f := range m.files {
		switch {
		case f.IsStaged() && !f.IsUnstaged():
			staged++
		case f.IsUnstaged() && !f.IsStaged():
			unstaged++
		case f.IsUntracked():
			untracked++
		default:
			if f.IsStaged() {
				staged++
			}
			if f.IsUnstaged() {
				unstaged++
			}
		}
	}
	parts := []string{}
	if staged > 0 {
		parts = append(parts, fmt.Sprintf("%d staged", staged))
	}
	if unstaged > 0 {
		parts = append(parts, fmt.Sprintf("%d unstaged", unstaged))
	}
	if untracked > 0 {
		parts = append(parts, fmt.Sprintf("%d untracked", untracked))
	}
	return strings.Join(parts, "  ")
}

// buildItems builds the display items for the scroll list, grouped by section.
func (m *FilesModel) buildItems() []string {
	var staged, unstaged, untracked []git.FileStatus
	for _, f := range m.files {
		switch {
		case f.IsConflicted():
			unstaged = append(unstaged, f)
		case f.IsUntracked():
			untracked = append(untracked, f)
		case f.IsStaged() && f.IsUnstaged():
			staged = append(staged, f)
			unstaged = append(unstaged, f)
		case f.IsStaged():
			staged = append(staged, f)
		default:
			unstaged = append(unstaged, f)
		}
	}

	var items []string

	if len(staged) > 0 {
		items = append(items, "── Staged ──────────────")
		for _, f := range staged {
			items = append(items, renderFileItem(f, true))
		}
	}
	if len(unstaged) > 0 {
		items = append(items, "── Unstaged ────────────")
		for _, f := range unstaged {
			items = append(items, renderFileItem(f, false))
		}
	}
	if len(untracked) > 0 {
		items = append(items, "── Untracked ───────────")
		for _, f := range untracked {
			items = append(items, renderFileItem(f, false))
		}
	}
	return items
}

func renderFileItem(f git.FileStatus, staged bool) string {
	badge := widgets.StatusBadge(f)
	name := filepath.Base(f.Path)
	dir := filepath.Dir(f.Path)
	if dir == "." {
		dir = ""
	} else {
		dir = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(dir+"/")
	}
	return fmt.Sprintf(" %s %s%s", badge, dir, name)
}

func isSectionHeader(item string) bool {
	return strings.HasPrefix(item, "──")
}
