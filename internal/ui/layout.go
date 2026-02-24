package ui

// PanelID identifies which panel is focused.
type PanelID int

const (
	PanelFiles PanelID = iota
	PanelBranches
	PanelCommits
	PanelDiff
	PanelCommitMsg
)

// numMainPanels is the number of panels in the main tab bar.
const numMainPanels = 3 // Files, Branches, Commits

// Layout holds computed dimensions for the current terminal size.
type Layout struct {
	TotalWidth  int
	TotalHeight int

	// Left column (Files + Branches + Commits)
	LeftWidth  int
	LeftHeight int

	// Right column (Diff)
	RightWidth  int
	RightHeight int

	// Header / footer heights
	HeaderHeight int
	FooterHeight int

	// Content heights (inside borders)
	PanelContentHeight int
	DiffContentHeight  int

	// Responsive mode
	IsNarrow    bool // < 100 cols: two panels, no diff side-by-side
	IsVeryNarrow bool // < 60 cols: single panel
}

// ComputeLayout calculates panel dimensions from terminal size.
func ComputeLayout(width, height int) Layout {
	l := Layout{
		TotalWidth:   width,
		TotalHeight:  height,
		HeaderHeight: 1, // single-line header
		FooterHeight: 1, // single-line status bar
	}

	usable := height - l.HeaderHeight - l.FooterHeight

	l.IsVeryNarrow = width < 60
	l.IsNarrow = width < 100

	switch {
	case l.IsVeryNarrow:
		// Single panel mode
		l.LeftWidth = width
		l.RightWidth = 0
	case l.IsNarrow:
		// Two-panel: left=40%, right=60%
		l.LeftWidth = width * 2 / 5
		l.RightWidth = width - l.LeftWidth
	default:
		// Three-panel: left=35%, right=65%
		l.LeftWidth = width * 35 / 100
		l.RightWidth = width - l.LeftWidth
	}

	l.LeftHeight = usable
	l.RightHeight = usable

	// Content height inside a bordered panel (2 lines for top+bottom border, 1 for title)
	l.PanelContentHeight = usable - 3
	l.DiffContentHeight = usable - 3

	return l
}
