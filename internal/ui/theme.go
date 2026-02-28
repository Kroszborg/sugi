package ui

import "github.com/charmbracelet/lipgloss"

// ── Color Palette (Catppuccin Mocha + enhancements) ───────────────────────────
const (
	// Backgrounds
	ColorBase    = lipgloss.Color("#11111b") // darkest — main bg
	ColorMantle  = lipgloss.Color("#181825") // header/footer bar
	ColorCrust   = lipgloss.Color("#1e1e2e") // panel bg
	ColorSurface = lipgloss.Color("#313244") // selected row bg
	ColorOverlay = lipgloss.Color("#45475a") // borders unfocused
	ColorMuted   = lipgloss.Color("#585b70") // dim text
	ColorSubtext = lipgloss.Color("#a6adc8") // secondary text
	ColorText    = lipgloss.Color("#cdd6f4") // primary text

	// Accents
	ColorBlue   = lipgloss.Color("#89b4fa") // blue
	ColorSky    = lipgloss.Color("#89dceb") // sky — focus highlight
	ColorTeal   = lipgloss.Color("#94e2d5") // teal — AI / success
	ColorGreen  = lipgloss.Color("#a6e3a1") // green — added
	ColorYellow = lipgloss.Color("#f9e2af") // yellow — warning
	ColorPeach  = lipgloss.Color("#fab387") // peach — section headers
	ColorOrange = lipgloss.Color("#fe640b") // orange — hashes
	ColorRed    = lipgloss.Color("#f38ba8") // red — deleted / error
	ColorPurple = lipgloss.Color("#cba6f7") // purple — author
	ColorPink   = lipgloss.Color("#f5c2e7") // pink — renamed
)

// Theme holds all Lipgloss styles used throughout the TUI.
type Theme struct {
	// Borders
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style

	// Text hierarchy
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Normal   lipgloss.Style
	Muted    lipgloss.Style
	Bold     lipgloss.Style

	// Selection
	Selected lipgloss.Style
	Cursor   lipgloss.Style

	// Git status badges (compact, icon-first)
	BadgeModified  lipgloss.Style
	BadgeAdded     lipgloss.Style
	BadgeDeleted   lipgloss.Style
	BadgeUntracked lipgloss.Style
	BadgeConflict  lipgloss.Style
	BadgeRenamed   lipgloss.Style

	// Diff
	DiffAdded   lipgloss.Style
	DiffRemoved lipgloss.Style
	DiffContext lipgloss.Style
	DiffHeader  lipgloss.Style
	DiffHunk    lipgloss.Style

	// Branch
	BranchCurrent lipgloss.Style
	BranchNormal  lipgloss.Style
	BranchRemote  lipgloss.Style

	// Commit
	CommitHash    lipgloss.Style
	CommitSubject lipgloss.Style
	CommitAuthor  lipgloss.Style
	CommitDate    lipgloss.Style
	CommitRef     lipgloss.Style

	// Status bar
	StatusBar   lipgloss.Style
	StatusKey   lipgloss.Style
	StatusValue lipgloss.Style
	StatusSep   lipgloss.Style

	// Help
	HelpKey   lipgloss.Style
	HelpValue lipgloss.Style
}

// DefaultTheme returns the polished dark theme.
var DefaultTheme = Theme{
	FocusedBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSky),

	UnfocusedBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOverlay),

	Title:    lipgloss.NewStyle().Foreground(ColorSky).Bold(true),
	Subtitle: lipgloss.NewStyle().Foreground(ColorSubtext),
	Normal:   lipgloss.NewStyle().Foreground(ColorText),
	Muted:    lipgloss.NewStyle().Foreground(ColorMuted),
	Bold:     lipgloss.NewStyle().Foreground(ColorText).Bold(true),

	Selected: lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorText),
	Cursor: lipgloss.NewStyle().
		Background(ColorSky).
		Foreground(ColorBase).
		Bold(true),

	BadgeModified:  lipgloss.NewStyle().Foreground(ColorYellow).Bold(true),
	BadgeAdded:     lipgloss.NewStyle().Foreground(ColorGreen).Bold(true),
	BadgeDeleted:   lipgloss.NewStyle().Foreground(ColorRed).Bold(true),
	BadgeUntracked: lipgloss.NewStyle().Foreground(ColorPeach).Bold(true),
	BadgeConflict:  lipgloss.NewStyle().Foreground(ColorRed).Bold(true),
	BadgeRenamed:   lipgloss.NewStyle().Foreground(ColorPink).Bold(true),

	DiffAdded:   lipgloss.NewStyle().Foreground(ColorGreen),
	DiffRemoved: lipgloss.NewStyle().Foreground(ColorRed),
	DiffContext: lipgloss.NewStyle().Foreground(ColorSubtext),
	DiffHeader:  lipgloss.NewStyle().Foreground(ColorBlue).Bold(true),
	DiffHunk:    lipgloss.NewStyle().Foreground(ColorTeal),

	BranchCurrent: lipgloss.NewStyle().Foreground(ColorGreen).Bold(true),
	BranchNormal:  lipgloss.NewStyle().Foreground(ColorText),
	BranchRemote:  lipgloss.NewStyle().Foreground(ColorSubtext),

	CommitHash:    lipgloss.NewStyle().Foreground(ColorPeach).Bold(true),
	CommitSubject: lipgloss.NewStyle().Foreground(ColorText),
	CommitAuthor:  lipgloss.NewStyle().Foreground(ColorPurple),
	CommitDate:    lipgloss.NewStyle().Foreground(ColorMuted),
	CommitRef:     lipgloss.NewStyle().Foreground(ColorYellow),

	StatusBar: lipgloss.NewStyle().
		Background(ColorMantle).
		Foreground(ColorText),
	StatusKey: lipgloss.NewStyle().
		Background(ColorMantle).
		Foreground(ColorSky).
		Bold(true),
	StatusValue: lipgloss.NewStyle().
		Background(ColorMantle).
		Foreground(ColorSubtext),
	StatusSep: lipgloss.NewStyle().
		Background(ColorMantle).
		Foreground(ColorOverlay),

	HelpKey:   lipgloss.NewStyle().Foreground(ColorSky).Bold(true),
	HelpValue: lipgloss.NewStyle().Foreground(ColorSubtext),
}
