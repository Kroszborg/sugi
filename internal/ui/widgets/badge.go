package widgets

import (
	"github.com/Kroszborg/sugi/internal/git"
	"github.com/charmbracelet/lipgloss"
)

// StatusBadge renders a colored single-character status badge for a file.
func StatusBadge(fs git.FileStatus) string {
	if fs.IsConflicted() {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Bold(true).Render("U")
	}
	if fs.IsUntracked() {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true).Render("?")
	}

	staged := stagedBadge(fs.Staged)
	unstaged := unstagedBadge(fs.Unstaged)

	if staged != "" && unstaged != "" {
		return staged + unstaged
	}
	if staged != "" {
		return staged + " "
	}
	if unstaged != "" {
		return " " + unstaged
	}
	return "  "
}

func stagedBadge(code git.StatusCode) string {
	switch code {
	case git.Modified:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Bold(true).Render("M")
	case git.Added:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true).Render("A")
	case git.Deleted:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Bold(true).Render("D")
	case git.Renamed:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true).Render("R")
	case git.Copied:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#89dceb")).Bold(true).Render("C")
	}
	return ""
}

func unstagedBadge(code git.StatusCode) string {
	switch code {
	case git.Modified:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render("m")
	case git.Deleted:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render("d")
	}
	return ""
}

// BranchBadge renders an indicator for a branch.
func BranchBadge(isCurrent bool) string {
	if isCurrent {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true).Render("*")
	}
	return " "
}

// AheadBehindBadge renders an ahead/behind indicator.
func AheadBehindBadge(ahead, behind int) string {
	if ahead == 0 && behind == 0 {
		return ""
	}
	s := ""
	if ahead > 0 {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Render("↑" + itoa(ahead))
	}
	if behind > 0 {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render("↓" + itoa(behind))
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 5)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
