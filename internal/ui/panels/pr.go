package panels

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/forge"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// PRModel displays pull requests with CI and review status.
type PRModel struct {
	prs    []forge.PullRequest
	list   widgets.ScrollList
	detail *forge.PullRequest // expanded PR view
	Width  int
	Height int

	showDetail bool

	// Create PR form
	creating     bool
	createTitle  string
	createBody   string
	createTarget string

	// Styles
	numberStyle    lipgloss.Style
	titleStyle     lipgloss.Style
	authorStyle    lipgloss.Style
	dateStyle      lipgloss.Style
	openStyle      lipgloss.Style
	draftStyle     lipgloss.Style
	mergedStyle    lipgloss.Style
	closedStyle    lipgloss.Style
	ciOKStyle      lipgloss.Style
	ciFailStyle    lipgloss.Style
	ciPendStyle    lipgloss.Style
	reviewOKStyle  lipgloss.Style
	reviewBadStyle lipgloss.Style
	emptyStyle     lipgloss.Style
	detailStyle    lipgloss.Style
}

// NewPRModel creates a PRModel.
func NewPRModel(width, height int) PRModel {
	return PRModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-4, width-4),

		numberStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true),
		titleStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
		authorStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")),
		dateStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")),
		openStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true),
		draftStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")),
		mergedStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true),
		closedStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
		ciOKStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
		ciFailStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
		ciPendStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")),
		reviewOKStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
		reviewBadStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")),
		emptyStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")),
		detailStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
	}
}

// SetPRs updates the PR list.
func (m *PRModel) SetPRs(prs []forge.PullRequest) {
	m.prs = prs
	m.list.SetItems(m.buildItems())
}

// CurrentPR returns the PR at the cursor.
func (m *PRModel) CurrentPR() *forge.PullRequest {
	if len(m.prs) == 0 || m.list.Cursor >= len(m.prs) {
		return nil
	}
	return &m.prs[m.list.Cursor]
}

// MoveUp moves cursor up.
func (m *PRModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves cursor down.
func (m *PRModel) MoveDown() { m.list.MoveDown() }

// ToggleDetail toggles the expanded PR detail view.
func (m *PRModel) ToggleDetail() {
	m.showDetail = !m.showDetail
	if m.showDetail {
		m.detail = m.CurrentPR()
	}
}

// View renders the PR panel.
func (m *PRModel) View() string {
	if m.showDetail && m.detail != nil {
		return m.renderDetail(m.detail)
	}

	if len(m.prs) == 0 {
		return m.emptyStyle.Render("  No pull requests found\n\n  n: create PR  P: push current branch first")
	}
	return m.list.View()
}

// StatusHint returns a context hint for the status bar.
func (m *PRModel) StatusHint() string {
	if len(m.prs) == 0 {
		return ""
	}
	return fmt.Sprintf("%d PRs", len(m.prs))
}

func (m *PRModel) buildItems() []string {
	items := make([]string, len(m.prs))
	for i, pr := range m.prs {
		items[i] = m.renderPRRow(pr, m.Width-4)
	}
	return items
}

func (m *PRModel) prStateBadge(pr forge.PullRequest) string {
	switch pr.State {
	case forge.PROpen:
		return m.openStyle.Render("● open  ")
	case forge.PRDraft:
		return m.draftStyle.Render("◌ draft ")
	case forge.PRMerged:
		return m.mergedStyle.Render("⬡ merged")
	case forge.PRClosed:
		return m.closedStyle.Render("✕ closed")
	}
	return ""
}

func (m *PRModel) prCIBadge(pr forge.PullRequest) string {
	switch pr.CI {
	case forge.CISuccess:
		return m.ciOKStyle.Render("✓")
	case forge.CIFailure, forge.CIError:
		return m.ciFailStyle.Render("✗")
	case forge.CIPending:
		return m.ciPendStyle.Render("⟳")
	}
	return " "
}

func (m *PRModel) prReviewBadge(pr forge.PullRequest) string {
	approved, changes := 0, 0
	for _, r := range pr.Reviews {
		switch r.State {
		case forge.ReviewApproved:
			approved++
		case forge.ReviewChangesRequested:
			changes++
		}
	}
	if approved > 0 && changes == 0 {
		return m.reviewOKStyle.Render(fmt.Sprintf("✓%d", approved))
	}
	if changes > 0 {
		return m.reviewBadStyle.Render(fmt.Sprintf("!%d", changes))
	}
	return ""
}

func (m *PRModel) renderPRRow(pr forge.PullRequest, width int) string {
	state := m.prStateBadge(pr)
	ci := m.prCIBadge(pr)
	review := m.prReviewBadge(pr)

	num := m.numberStyle.Render(fmt.Sprintf("#%-4d", pr.Number))
	author := m.authorStyle.Render(pr.Author)
	date := m.dateStyle.Render(relativeDate(pr.CreatedAt))

	// Title truncated
	metaW := 8 + len(pr.Author) + 12 + 10
	titleW := width - metaW
	if titleW < 10 {
		titleW = 10
	}
	title := pr.Title
	if len(title) > titleW {
		title = title[:titleW-1] + "…"
	}
	titleStr := m.titleStyle.Render(title)

	return fmt.Sprintf(" %s %s %s %s  %s %s  %s",
		state, num, titleStr, ci, review, author, date)
}

func (m *PRModel) renderDetail(pr *forge.PullRequest) string {
	var sb strings.Builder

	// Header
	sb.WriteString(m.numberStyle.Render(fmt.Sprintf("#%d ", pr.Number)))
	sb.WriteString(m.titleStyle.Render(pr.Title) + "\n\n")

	// Meta
	stateStr := string(pr.State)
	sb.WriteString(m.authorStyle.Render("Author: "+pr.Author) + "  ")
	sb.WriteString(m.dateStyle.Render("State: "+stateStr) + "  ")
	sb.WriteString(m.dateStyle.Render("Created: "+relativeDate(pr.CreatedAt)) + "\n")

	// Branches
	sb.WriteString(m.dateStyle.Render(
		fmt.Sprintf("Branch: %s → %s\n", pr.HeadBranch, pr.BaseBranch)))

	// CI status
	switch pr.CI {
	case forge.CISuccess:
		sb.WriteString(m.ciOKStyle.Render("CI: ✓ passing") + "\n")
	case forge.CIFailure:
		sb.WriteString(m.ciFailStyle.Render("CI: ✗ failing") + "\n")
	case forge.CIPending:
		sb.WriteString(m.ciPendStyle.Render("CI: ⟳ running") + "\n")
	}

	// Reviews
	if len(pr.Reviews) > 0 {
		sb.WriteString("\nReviews:\n")
		for _, r := range pr.Reviews {
			icon := "○"
			switch r.State {
			case forge.ReviewApproved:
				icon = m.reviewOKStyle.Render("✓")
			case forge.ReviewChangesRequested:
				icon = m.reviewBadStyle.Render("!")
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n", icon, m.authorStyle.Render(r.Author)))
		}
	}

	// Body
	if pr.Body != "" {
		sb.WriteString("\n")
		body := pr.Body
		maxBodyLines := m.Height - 15
		if maxBodyLines < 3 {
			maxBodyLines = 3
		}
		lines := strings.Split(body, "\n")
		if len(lines) > maxBodyLines {
			lines = lines[:maxBodyLines]
			lines = append(lines, m.dateStyle.Render("… (truncated)"))
		}
		sb.WriteString(m.detailStyle.Render(strings.Join(lines, "\n")))
	}

	// URL
	sb.WriteString("\n\n" + m.dateStyle.Render(pr.URL))
	sb.WriteString("\n\n" + m.dateStyle.Render("[m] merge  [x] close  [esc] back to list"))

	return sb.String()
}

// CIBadge renders a CI status badge for inline display.
func CIBadge(result forge.CIResult) string {
	switch result {
	case forge.CISuccess:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Render("✓")
	case forge.CIFailure, forge.CIError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render("✗")
	case forge.CIPending:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render("⟳")
	default:
		return " "
	}
}

// PRBadge renders a PR number badge for inline display (e.g. in branch list).
func PRBadge(number int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89b4fa")).
		Render(fmt.Sprintf("PR#%d", number))
}

// ListCursor returns the current scroll list cursor position.
func (m *PRModel) ListCursor() int { return m.list.Cursor }

// SetListCursor sets the scroll list cursor position.
func (m *PRModel) SetListCursor(n int) { m.list.Cursor = n }
