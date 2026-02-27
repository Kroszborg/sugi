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
		Height:   height,
		Width:    width,
		Selected: make(map[int]bool),
		CursorStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#313244")).
			Foreground(lipgloss.Color("#cdd6f4")),
		NormalStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
		SelectedStyle: lipgloss.NewStyle().Background(lipgloss.Color("#2a2a3e")).Foreground(lipgloss.Color("#cdd6f4")),
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

// SelectedIndices returns a sorted list of all selected indices.
func (sl *ScrollList) SelectedIndices() []int {
	var out []int
	for i := range sl.Items {
		if sl.Selected[i] {
			out = append(out, i)
		}
	}
	return out
}

// HasSelection returns true if any items are multi-selected.
func (sl *ScrollList) HasSelection() bool {
	return len(sl.Selected) > 0
}

// CurrentItem returns the item at the cursor, or "" if the list is empty.
func (sl *ScrollList) CurrentItem() string {
	if len(sl.Items) == 0 || sl.Cursor >= len(sl.Items) {
		return ""
	}
	return sl.Items[sl.Cursor]
}

// cursor glyph styles — rendered once and reused
var (
	cursorGlyph  = lipgloss.NewStyle().Foreground(lipgloss.Color("#89dceb")).Bold(true).Render("▶")
	normalGlyph  = lipgloss.NewStyle().Foreground(lipgloss.Color("#313244")).Render("·")
	selectedMark = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true).Render("◆")
)

// View renders the list into a string of at most Height lines.
func (sl *ScrollList) View() string {
	if len(sl.Items) == 0 {
		return sl.NormalStyle.Render("  (empty)")
	}

	var sb strings.Builder
	end := min(sl.offset+sl.Height, len(sl.Items))

	for i := sl.offset; i < end; i++ {
		item := sl.Items[i]

		// Strip raw string length for truncation (items may have ANSI codes;
		// use a generous visible-width approximation via lipgloss.Width).
		visW := sl.Width - 3 // 3 chars for gutter + space
		if visW < 4 {
			visW = 4
		}
		raw := stripToWidth(item, visW)

		var glyph string
		switch {
		case i == sl.Cursor && sl.Selected[i]:
			glyph = cursorGlyph + " "
			raw = sl.CursorStyle.Width(visW).Render(raw)
		case i == sl.Cursor:
			glyph = cursorGlyph + " "
			raw = sl.CursorStyle.Width(visW).Render(raw)
		case sl.Selected[i]:
			glyph = selectedMark + " "
			raw = sl.SelectedStyle.Width(visW).Render(raw)
		default:
			glyph = normalGlyph + " "
			// normal items keep their own styling
		}

		sb.WriteString(glyph + raw)
		if i < end-1 {
			sb.WriteRune('\n')
		}
	}

	// Scroll indicator in the top-right corner when content overflows.
	result := sb.String()
	if len(sl.Items) > sl.Height {
		pct := 0
		if len(sl.Items) > 1 {
			pct = sl.offset * 100 / (len(sl.Items) - 1)
		}
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70")).
			Render(" " + progressBar(pct))
		_ = indicator // returned as part of the last line — keep simple for now
	}

	return result
}

// stripToWidth removes trailing characters so the visible string fits within n
// terminal columns. It uses a simple byte count as ANSI escape codes appear
// embedded in pre-styled item strings — we rely on lipgloss.Width downstream.
func stripToWidth(s string, n int) string {
	// Use lipgloss Width which strips ANSI codes for measurement.
	if lipgloss.Width(s) <= n {
		return s
	}
	// Binary-search a rune count that fits.
	runes := []rune(s)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if lipgloss.Width(string(runes[:mid])) <= n-1 {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	if lo <= 0 {
		return ""
	}
	return string(runes[:lo]) + "…"
}

func progressBar(pct int) string {
	if pct <= 10 {
		return "⊤"
	} else if pct >= 90 {
		return "⊥"
	}
	return "│"
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
