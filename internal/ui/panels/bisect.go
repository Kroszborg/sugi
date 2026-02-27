package panels

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// BisectModel displays the git bisect session status and log.
type BisectModel struct {
	Status git.BisectStatus
	list   widgets.ScrollList
	Width  int
	Height int
}

// NewBisectModel creates a BisectModel.
func NewBisectModel(width, height int) BisectModel {
	return BisectModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-6, width-4),
	}
}

// SetStatus updates the bisect status.
func (m *BisectModel) SetStatus(status git.BisectStatus) {
	m.Status = status
	m.list.SetItems(status.Log)
}

// MoveUp moves the log cursor up.
func (m *BisectModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the log cursor down.
func (m *BisectModel) MoveDown() { m.list.MoveDown() }

// View renders the bisect panel.
func (m *BisectModel) View() string {
	if !m.Status.InProgress {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).
			Render("  No bisect session in progress.\n\n  Press [s] to start, then [b]ad and [g]ood to narrow down.")
	}

	// Status header
	hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true)
	goodStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true)
	badStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70"))

	testing := ""
	if m.Status.CurrentHash != "" {
		testing = fmt.Sprintf("  %s %s\n",
			labelStyle.Render("Testing:"),
			hashStyle.Render(m.Status.CurrentHash),
		)
	}

	counts := fmt.Sprintf("  %s %s   %s %s\n",
		goodStyle.Render(fmt.Sprintf("✓ %d good", m.Status.GoodCount)),
		labelStyle.Render("·"),
		badStyle.Render(fmt.Sprintf("✗ %d bad", m.Status.BadCount)),
		labelStyle.Render("commits marked"),
	)

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a"))
	sep := sepStyle.Render(strings.Repeat("─", m.Width-6)) + "\n"

	logTitle := labelStyle.Render("  Bisect log:\n")

	return testing + counts + sep + logTitle + m.list.View()
}
