package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
const (
	ColorBase    = lipgloss.Color("#1e1e2e") // dark background
	ColorSurface = lipgloss.Color("#313244")
	ColorOverlay = lipgloss.Color("#45475a")
	ColorMuted   = lipgloss.Color("#585b70")
	ColorText    = lipgloss.Color("#cdd6f4")
	ColorSubtext = lipgloss.Color("#a6adc8")

	ColorBlue    = lipgloss.Color("#89b4fa")
	ColorGreen   = lipgloss.Color("#a6e3a1")
	ColorRed     = lipgloss.Color("#f38ba8")
	ColorYellow  = lipgloss.Color("#f9e2af")
	ColorOrange  = lipgloss.Color("#fab387")
	ColorPurple  = lipgloss.Color("#cba6f7")
	ColorTeal    = lipgloss.Color("#94e2d5")
	ColorSky     = lipgloss.Color("#89dceb")
)

// Theme holds all Lipgloss styles used throughout the TUI.
type Theme struct {
	// Borders
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style

	// Text
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Normal   lipgloss.Style
	Muted    lipgloss.Style
	Bold     lipgloss.Style

	// Selection
	Selected   lipgloss.Style
	Cursor     lipgloss.Style

	// Git status badges
	BadgeModified  lipgloss.Style
	BadgeAdded     lipgloss.Style
	BadgeDeleted   lipgloss.Style
	BadgeUntracked lipgloss.Style
	BadgeConflict  lipgloss.Style
	BadgeRenamed   lipgloss.Style

	// Diff colors
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
	StatusBar     lipgloss.Style
	StatusKey     lipgloss.Style
	StatusValue   lipgloss.Style
	StatusSep     lipgloss.Style

	// Help
	HelpKey   lipgloss.Style
	HelpValue lipgloss.Style
}

// DefaultTheme returns the default dark theme (catppuccin-inspired).
var DefaultTheme = Theme{
	FocusedBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBlue),

	UnfocusedBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted),

	Title:    lipgloss.NewStyle().Foreground(ColorBlue).Bold(true),
	Subtitle: lipgloss.NewStyle().Foreground(ColorSubtext),
	Normal:   lipgloss.NewStyle().Foreground(ColorText),
	Muted:    lipgloss.NewStyle().Foreground(ColorMuted),
	Bold:     lipgloss.NewStyle().Foreground(ColorText).Bold(true),

	Selected: lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorText),
	Cursor: lipgloss.NewStyle().
		Background(ColorBlue).
		Foreground(ColorBase).
		Bold(true),

	BadgeModified:  lipgloss.NewStyle().Foreground(ColorYellow).Bold(true),
	BadgeAdded:     lipgloss.NewStyle().Foreground(ColorGreen).Bold(true),
	BadgeDeleted:   lipgloss.NewStyle().Foreground(ColorRed).Bold(true),
	BadgeUntracked: lipgloss.NewStyle().Foreground(ColorOrange).Bold(true),
	BadgeConflict:  lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Blink(true),
	BadgeRenamed:   lipgloss.NewStyle().Foreground(ColorPurple).Bold(true),

	DiffAdded:   lipgloss.NewStyle().Foreground(ColorGreen),
	DiffRemoved: lipgloss.NewStyle().Foreground(ColorRed),
	DiffContext: lipgloss.NewStyle().Foreground(ColorText),
	DiffHeader:  lipgloss.NewStyle().Foreground(ColorBlue).Bold(true),
	DiffHunk:    lipgloss.NewStyle().Foreground(ColorTeal),

	BranchCurrent: lipgloss.NewStyle().Foreground(ColorGreen).Bold(true),
	BranchNormal:  lipgloss.NewStyle().Foreground(ColorText),
	BranchRemote:  lipgloss.NewStyle().Foreground(ColorSubtext),

	CommitHash:    lipgloss.NewStyle().Foreground(ColorOrange).Bold(true),
	CommitSubject: lipgloss.NewStyle().Foreground(ColorText),
	CommitAuthor:  lipgloss.NewStyle().Foreground(ColorPurple),
	CommitDate:    lipgloss.NewStyle().Foreground(ColorMuted),
	CommitRef:     lipgloss.NewStyle().Foreground(ColorYellow),

	StatusBar: lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorText),
	StatusKey: lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorBlue).
		Bold(true),
	StatusValue: lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorSubtext),
	StatusSep: lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorOverlay),

	HelpKey:   lipgloss.NewStyle().Foreground(ColorBlue).Bold(true),
	HelpValue: lipgloss.NewStyle().Foreground(ColorSubtext),
}
