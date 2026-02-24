package widgets

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ScrollList is a navigable, scrollable list of string items.
type ScrollList struct {
	Items    []string
	Cursor   int
	offset   int // first visible item
	Height   int // visible rows
	Width    int
	Selected map[int]bool // multi-select

	// Styles
	CursorStyle   lipgloss.Style
	NormalStyle   lipgloss.Style
	SelectedStyle lipgloss.Style
}

// NewScrollList creates a ScrollList with sensible defaults.
func NewScrollList(height, width int) ScrollList {
	return ScrollList{
		Height:        height,
		Width:         width,
		Selected:      make(map[int]bool),
		CursorStyle:   lipgloss.NewStyle().Background(lipgloss.Color("#89b4fa")).Foreground(lipgloss.Color("#1e1e2e")).Bold(true),
		NormalStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
		SelectedStyle: lipgloss.NewStyle().Background(lipgloss.Color("#313244")).Foreground(lipgloss.Color("#cdd6f4")),
	}
}

// SetItems replaces the list contents, resetting cursor if out of range.
func (sl *ScrollList) SetItems(items []string) {
	sl.Items = items
	if sl.Cursor >= len(items) {
		sl.Cursor = max(0, len(items)-1)
	}
	sl.clampOffset()
}

// MoveUp moves the cursor up by one.
func (sl *ScrollList) MoveUp() {
	if sl.Cursor > 0 {
		sl.Cursor--
		sl.clampOffset()
	}
}

// MoveDown moves the cursor down by one.
func (sl *ScrollList) MoveDown() {
	if sl.Cursor < len(sl.Items)-1 {
		sl.Cursor++
		sl.clampOffset()
	}
}

// PageUp moves the cursor up by one page.
func (sl *ScrollList) PageUp() {
	sl.Cursor -= sl.Height
	if sl.Cursor < 0 {
		sl.Cursor = 0
	}
	sl.clampOffset()
}

// PageDown moves the cursor down by one page.
func (sl *ScrollList) PageDown() {
	sl.Cursor += sl.Height
	if sl.Cursor >= len(sl.Items) {
		sl.Cursor = max(0, len(sl.Items)-1)
	}
	sl.clampOffset()
}

// Top moves the cursor to the first item.
func (sl *ScrollList) Top() {
	sl.Cursor = 0
	sl.offset = 0
}

// Bottom moves the cursor to the last item.
func (sl *ScrollList) Bottom() {
	sl.Cursor = max(0, len(sl.Items)-1)
	sl.clampOffset()
}

// ToggleSelect toggles multi-selection of the current item.
func (sl *ScrollList) ToggleSelect() {
	if sl.Selected[sl.Cursor] {
		delete(sl.Selected, sl.Cursor)
	} else {
		sl.Selected[sl.Cursor] = true
	}
}

// ClearSelected clears all multi-selections.
func (sl *ScrollList) ClearSelected() {
	sl.Selected = make(map[int]bool)
}

// CurrentItem returns the item at the cursor, or "" if the list is empty.
func (sl *ScrollList) CurrentItem() string {
	if len(sl.Items) == 0 || sl.Cursor >= len(sl.Items) {
		return ""
	}
	return sl.Items[sl.Cursor]
}

// View renders the list into a string of at most Height lines.
func (sl *ScrollList) View() string {
	if len(sl.Items) == 0 {
		return sl.NormalStyle.Render("  (empty)")
	}

	var sb strings.Builder
	end := min(sl.offset+sl.Height, len(sl.Items))

	for i := sl.offset; i < end; i++ {
		item := sl.Items[i]
		// Truncate to width
		if sl.Width > 0 && len(item) > sl.Width {
			item = item[:sl.Width-1] + "…"
		}

		switch {
		case i == sl.Cursor:
			sb.WriteString(sl.CursorStyle.Width(sl.Width).Render(item))
		case sl.Selected[i]:
			sb.WriteString(sl.SelectedStyle.Width(sl.Width).Render(item))
		default:
			sb.WriteString(sl.NormalStyle.Width(sl.Width).Render(item))
		}

		if i < end-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// clampOffset ensures the cursor is visible within the scrolled window.
func (sl *ScrollList) clampOffset() {
	if sl.Cursor < sl.offset {
		sl.offset = sl.Cursor
	}
	if sl.Cursor >= sl.offset+sl.Height {
		sl.offset = sl.Cursor - sl.Height + 1
	}
	if sl.offset < 0 {
		sl.offset = 0
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
