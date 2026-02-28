package ui

import "github.com/charmbracelet/lipgloss"

// ── Color Palette — Deep Noir (inspired by Linear.app × deep space) ───────────
//
// 3-tier accent discipline:
//   L1 Neutral  — 6 surface depths (background → text)
//   L2 Accent   — ONE interactive color: electric violet #7c6dfa
//   L3 Semantic — green (added/ok) · red (removed/err) · amber (warn)
const (
	// Neutral surfaces — darkest to brightest
	ColorBase    = lipgloss.Color("#08080f") // deepest bg
	ColorMantle  = lipgloss.Color("#0e0e1a") // header/footer bar
	ColorCrust   = lipgloss.Color("#13131e") // panel bg
	ColorSurface = lipgloss.Color("#1a1a2a") // selected row bg
	ColorOverlay = lipgloss.Color("#252538") // borders unfocused
	ColorMuted   = lipgloss.Color("#3d3d5c") // truly dim
	ColorSubtext = lipgloss.Color("#7878a0") // secondary text
	ColorText    = lipgloss.Color("#d8d8ee") // primary text

	// L2 — ONE primary accent: electric violet
	ColorAccent     = lipgloss.Color("#7c6dfa") // focus / active / interactive
	ColorAccentDim  = lipgloss.Color("#4a3faa") // unfocused accent

	// L3 Semantic — git / diff / status
	ColorGreen  = lipgloss.Color("#3ecf8e") // added · success · current branch
	ColorRed    = lipgloss.Color("#e05454") // removed · error · conflict
	ColorAmber  = lipgloss.Color("#d4a017") // warning · staged · rebase
	ColorBlue   = lipgloss.Color("#4d9de0") // branches · info · remote
	ColorPurple = lipgloss.Color("#a87efb") // author · merge · reword
	ColorPeach  = lipgloss.Color("#e8835c") // commit hash · section headers
	ColorTeal   = lipgloss.Color("#2ec4b6") // AI · teal accent
	ColorSky    = lipgloss.Color("#7c6dfa") // alias → same as Accent
	ColorYellow = lipgloss.Color("#d4a017") // alias → same as Amber
	ColorOrange = lipgloss.Color("#e8835c") // alias → same as Peach
	ColorPink   = lipgloss.Color("#e879a8") // renamed files
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

// DefaultTheme returns the noir dark theme.
var DefaultTheme = Theme{
	FocusedBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent),

	UnfocusedBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOverlay),

	Title:    lipgloss.NewStyle().Foreground(ColorAccent).Bold(true),
	Subtitle: lipgloss.NewStyle().Foreground(ColorSubtext),
	Normal:   lipgloss.NewStyle().Foreground(ColorText),
	Muted:    lipgloss.NewStyle().Foreground(ColorMuted),
	Bold:     lipgloss.NewStyle().Foreground(ColorText).Bold(true),

	Selected: lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorText),
	Cursor: lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorBase).
		Bold(true),

	BadgeModified:  lipgloss.NewStyle().Foreground(ColorAmber).Bold(true),
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
	CommitRef:     lipgloss.NewStyle().Foreground(ColorAmber),

	StatusBar: lipgloss.NewStyle().
		Background(ColorMantle).
		Foreground(ColorText),
	StatusKey: lipgloss.NewStyle().
		Background(ColorMantle).
		Foreground(ColorAccent).
		Bold(true),
	StatusValue: lipgloss.NewStyle().
		Background(ColorMantle).
		Foreground(ColorSubtext),
	StatusSep: lipgloss.NewStyle().
		Background(ColorMantle).
		Foreground(ColorOverlay),

	HelpKey:   lipgloss.NewStyle().Foreground(ColorAccent).Bold(true),
	HelpValue: lipgloss.NewStyle().Foreground(ColorSubtext),
}
