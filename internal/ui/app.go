package ui

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/panels"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Messages ---

// GitStatusMsg carries a fresh git status result.
type GitStatusMsg struct {
	Files []git.FileStatus
	Err   error
}

// GitBranchesMsg carries a fresh branches list.
type GitBranchesMsg struct {
	Branches []git.Branch
	Err      error
}

// GitCommitsMsg carries a fresh commit log.
type GitCommitsMsg struct {
	Commits []git.Commit
	Err     error
}

// GitDiffMsg carries diff hunks for a specific file.
type GitDiffMsg struct {
	Hunks  []git.DiffHunk
	Staged bool
	Err    error
}

// GitOperationMsg reports the result of a mutating git operation.
type GitOperationMsg struct {
	Op  string
	Err error
}

// StatusMsg sets the status bar message.
type StatusMsg struct {
	Text  string
	IsErr bool
}

// --- App Mode ---

// Mode describes the current input mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeCommit
	ModeHelp
	ModeNewBranch
)

// --- Model ---

// Model is the root Bubbletea model.
type Model struct {
	repo   *git.Client
	keymap KeyMap

	// Panels
	files     panels.FilesModel
	branches  panels.BranchModel
	commits   panels.CommitModel
	diff      panels.DiffModel
	commitMsg panels.CommitMsgModel

	// Focus
	focused PanelID

	// Mode
	mode Mode

	// Diff state
	diffStaged bool

	// Terminal dimensions
	width  int
	height int

	// Status bar
	statusBar widgets.StatusBar
	statusMsg string
	statusErr bool

	// Loading state
	loading    bool
	loadingMsg string

	// Theme
	theme Theme
}

// New creates the root application model for the given repo.
func New(repo *git.Client) Model {
	m := Model{
		repo:   repo,
		keymap: DefaultKeyMap(),
		theme:  DefaultTheme,
		width:  80,
		height: 24,
	}
	m.rebuildPanels()
	return m
}

// Init starts initial data loading.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadStatus(),
		m.loadBranches(),
		m.loadCommits(),
	)
}

// Update handles all messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildPanels()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case GitStatusMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Error loading status: " + msg.Err.Error()
			m.statusErr = true
		} else {
			m.files.SetFiles(msg.Files)
			m.statusMsg = ""
			m.statusErr = false
		}
		return m, nil

	case GitBranchesMsg:
		if msg.Err == nil {
			m.branches.SetBranches(msg.Branches)
		}
		return m, nil

	case GitCommitsMsg:
		if msg.Err == nil {
			m.commits.SetCommits(msg.Commits)
		}
		return m, nil

	case GitDiffMsg:
		if msg.Err == nil {
			m.diff.SetHunks(msg.Hunks)
		}
		return m, nil

	case GitOperationMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = fmt.Sprintf("%s failed: %s", msg.Op, msg.Err.Error())
			m.statusErr = true
		} else {
			m.statusMsg = msg.Op + " succeeded"
			m.statusErr = false
		}
		// Refresh after any mutating operation
		return m, tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits())

	case StatusMsg:
		m.statusMsg = msg.Text
		m.statusErr = msg.IsErr
		return m, nil
	}

	return m, nil
}

// View renders the full TUI.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}

	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// --- Key handling ---

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// CommitMsg mode captures most keys
	if m.mode == ModeCommit {
		return m.handleCommitKey(msg)
	}

	// New branch modal
	if m.mode == ModeNewBranch {
		return m.handleNewBranchKey(msg)
	}

	switch {
	case msg.String() == "q", msg.String() == "ctrl+c":
		return m, tea.Quit

	case msg.String() == "tab":
		m.focusNext()
		return m, m.loadDiffForCursor()

	case msg.String() == "shift+tab":
		m.focusPrev()
		return m, m.loadDiffForCursor()

	case msg.String() == "1":
		m.focused = PanelFiles
		return m, m.loadDiffForCursor()

	case msg.String() == "2":
		m.focused = PanelBranches
		return m, nil

	case msg.String() == "3":
		m.focused = PanelCommits
		return m, m.loadDiffForCursor()

	case msg.String() == "r":
		m.loading = true
		m.loadingMsg = "Refreshing…"
		return m, tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits())

	case msg.String() == "?":
		m.mode = ModeHelp
		return m, nil

	case msg.String() == "esc":
		if m.mode != ModeNormal {
			m.mode = ModeNormal
		}
		return m, nil
	}

	// Panel-specific keys
	switch m.focused {
	case PanelFiles:
		return m.handleFilesKey(msg)
	case PanelBranches:
		return m.handleBranchesKey(msg)
	case PanelCommits:
		return m.handleCommitsKey(msg)
	case PanelDiff:
		return m.handleDiffKey(msg)
	}

	return m, nil
}

func (m Model) handleFilesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.files.MoveDown()
		return m, m.loadDiffForCursor()
	case "k", "up":
		m.files.MoveUp()
		return m, m.loadDiffForCursor()
	case " ":
		return m, m.toggleStage()
	case "a":
		return m, m.stageAll()
	case "c":
		m.mode = ModeCommit
		m.commitMsg.Focus()
		return m, nil
	case "P":
		return m, m.push()
	case "p":
		return m, m.pull()
	case "f":
		return m, m.fetch()
	case "S":
		m.diffStaged = !m.diffStaged
		return m, m.loadDiffForCursor()
	}
	return m, nil
}

func (m Model) handleBranchesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.branches.MoveDown()
		return m, nil
	case "k", "up":
		m.branches.MoveUp()
		return m, nil
	case "enter":
		b := m.branches.CurrentBranch()
		if b != nil && !b.IsCurrent {
			return m, m.checkout(b.Name)
		}
	case "n":
		m.mode = ModeNewBranch
		m.branches.ShowNewBranchModal()
		return m, nil
	case "D":
		b := m.branches.CurrentBranch()
		if b != nil && !b.IsCurrent {
			return m, m.deleteBranch(b.Name)
		}
	}
	return m, nil
}

func (m Model) handleCommitsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.commits.MoveDown()
		return m, m.loadCommitDiff()
	case "k", "up":
		m.commits.MoveUp()
		return m, m.loadCommitDiff()
	}
	return m, nil
}

func (m Model) handleDiffKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.diff.ScrollDown()
	case "k", "up":
		m.diff.ScrollUp()
	case "ctrl+f", "pgdown":
		m.diff.PageDown()
	case "ctrl+b", "pgup":
		m.diff.PageUp()
	}
	return m, nil
}

func (m Model) handleCommitKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.commitMsg.Blur()
		m.commitMsg.Reset()
		return m, nil
	case "ctrl+s":
		text := m.commitMsg.Value()
		if text == "" {
			return m, func() tea.Msg { return StatusMsg{Text: "Commit message cannot be empty", IsErr: true} }
		}
		m.mode = ModeNormal
		m.commitMsg.Blur()
		msgText := text
		m.commitMsg.Reset()
		return m, m.doCommit(msgText)
	}
	// Pass key to textarea
	cmd := m.commitMsg.Update(msg)
	return m, cmd
}

func (m Model) handleNewBranchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.branches.HideModal()
		return m, nil
	case "enter":
		name := m.branches.ModalInput()
		m.mode = ModeNormal
		m.branches.HideModal()
		if name != "" {
			return m, m.createBranch(name)
		}
		return m, nil
	}
	// Pass key to modal input
	cmd := m.branches.UpdateModalInput(msg)
	return m, cmd
}

// --- Git commands as tea.Cmd ---

func (m Model) loadStatus() tea.Cmd {
	return func() tea.Msg {
		files, err := m.repo.Status()
		return GitStatusMsg{Files: files, Err: err}
	}
}

func (m Model) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.repo.Branches()
		return GitBranchesMsg{Branches: branches, Err: err}
	}
}

func (m Model) loadCommits() tea.Cmd {
	return func() tea.Msg {
		commits, err := m.repo.Log(100)
		return GitCommitsMsg{Commits: commits, Err: err}
	}
}

func (m Model) loadDiffForCursor() tea.Cmd {
	switch m.focused {
	case PanelFiles:
		f := m.files.CurrentFile()
		if f == nil {
			return func() tea.Msg { return GitDiffMsg{} }
		}
		path := f.Path
		staged := m.diffStaged
		return func() tea.Msg {
			hunks, err := m.repo.Diff(path, staged)
			return GitDiffMsg{Hunks: hunks, Staged: staged, Err: err}
		}
	case PanelCommits:
		return m.loadCommitDiff()
	}
	return nil
}

func (m Model) loadCommitDiff() tea.Cmd {
	c := m.commits.CurrentCommit()
	if c == nil {
		return nil
	}
	hash := c.Hash
	return func() tea.Msg {
		fds, err := m.repo.DiffCommit(hash)
		if err != nil {
			return GitDiffMsg{Err: err}
		}
		var allHunks []git.DiffHunk
		for _, fd := range fds {
			allHunks = append(allHunks, fd.Hunks...)
		}
		return GitDiffMsg{Hunks: allHunks}
	}
}

func (m Model) toggleStage() tea.Cmd {
	f := m.files.CurrentFile()
	if f == nil {
		return nil
	}
	path := f.Path
	isStaged := f.IsStaged()
	return func() tea.Msg {
		var err error
		if isStaged {
			err = m.repo.Unstage(path)
		} else {
			err = m.repo.Stage(path)
		}
		op := "Stage"
		if isStaged {
			op = "Unstage"
		}
		return GitOperationMsg{Op: op, Err: err}
	}
}

func (m Model) stageAll() tea.Cmd {
	return func() tea.Msg {
		err := m.repo.StageAll()
		return GitOperationMsg{Op: "Stage all", Err: err}
	}
}

func (m Model) doCommit(message string) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.Commit(message)
		return GitOperationMsg{Op: "Commit", Err: err}
	}
}

func (m Model) push() tea.Cmd {
	return func() tea.Msg {
		err := m.repo.Push()
		return GitOperationMsg{Op: "Push", Err: err}
	}
}

func (m Model) pull() tea.Cmd {
	return func() tea.Msg {
		err := m.repo.Pull()
		return GitOperationMsg{Op: "Pull", Err: err}
	}
}

func (m Model) fetch() tea.Cmd {
	return func() tea.Msg {
		err := m.repo.Fetch()
		return GitOperationMsg{Op: "Fetch", Err: err}
	}
}

func (m Model) checkout(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.Checkout(name)
		return GitOperationMsg{Op: "Checkout " + name, Err: err}
	}
}

func (m Model) createBranch(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.CreateBranch(name)
		return GitOperationMsg{Op: "Create branch " + name, Err: err}
	}
}

func (m Model) deleteBranch(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.DeleteBranch(name)
		return GitOperationMsg{Op: "Delete branch " + name, Err: err}
	}
}

// --- View helpers ---

func (m Model) renderHeader() string {
	repoName := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89b4fa")).
		Bold(true).
		Render("  " + m.repo.RepoName())

	branch, _ := m.repo.CurrentBranch()
	branchStr := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a6e3a1")).
		Render("  " + branch)

	loading := ""
	if m.loading {
		loading = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render("  " + m.loadingMsg)
	}

	right := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#585b70")).
		Render("sugi")

	left := repoName + branchStr + loading
	pad := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#181825")).
		Width(m.width).
		Render(left + strings.Repeat(" ", pad) + right)
}

func (m Model) renderBody() string {
	layout := ComputeLayout(m.width, m.height)
	panelH := layout.LeftHeight
	leftW := layout.LeftWidth
	rightW := layout.RightWidth

	// Render panels based on focus and layout
	var left, right string

	if layout.IsVeryNarrow {
		// Single panel
		switch m.focused {
		case PanelFiles:
			left = m.renderFilesPanel(leftW, panelH)
		case PanelBranches:
			left = m.renderBranchesPanel(leftW, panelH)
		case PanelCommits:
			left = m.renderCommitsPanel(leftW, panelH)
		case PanelDiff:
			left = m.renderDiffPanel(leftW, panelH)
		}
		return left
	}

	// Left column: stacked panels
	// Show focused left panel + all three tab options
	tabH := panelH / 3

	filesPanel := m.renderFilesPanel(leftW, tabH)
	branchesPanel := m.renderBranchesPanel(leftW, tabH)
	commitsPanel := m.renderCommitsPanel(leftW, panelH-2*tabH)

	left = lipgloss.JoinVertical(lipgloss.Left, filesPanel, branchesPanel, commitsPanel)

	if !layout.IsNarrow {
		right = m.renderDiffPanel(rightW, panelH)
	}

	if right == "" {
		return left
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) renderFilesPanel(width, height int) string {
	focused := m.focused == PanelFiles && m.mode == ModeNormal
	borderColor := lipgloss.Color("#585b70")
	titleColor := lipgloss.Color("#a6adc8")
	if focused {
		borderColor = lipgloss.Color("#89b4fa")
		titleColor = lipgloss.Color("#89b4fa")
	}

	inner := m.files.View()
	if m.mode == ModeCommit {
		inner = m.commitMsg.View()
	}

	title := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(focused).
		Render("FILES [1]")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2).
		Render(title + "\n" + inner)
}

func (m Model) renderBranchesPanel(width, height int) string {
	focused := m.focused == PanelBranches && m.mode == ModeNormal
	borderColor := lipgloss.Color("#585b70")
	titleColor := lipgloss.Color("#a6adc8")
	if focused {
		borderColor = lipgloss.Color("#89b4fa")
		titleColor = lipgloss.Color("#89b4fa")
	}

	title := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(focused).
		Render("BRANCHES [2]")

	inner := m.branches.View()

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2).
		Render(title + "\n" + inner)

	// Overlay modal if visible
	if m.branches.IsModalVisible() {
		modalView := m.branches.ModalView()
		_ = modalView // TODO: proper overlay positioning
	}

	return box
}

func (m Model) renderCommitsPanel(width, height int) string {
	focused := m.focused == PanelCommits && m.mode == ModeNormal
	borderColor := lipgloss.Color("#585b70")
	titleColor := lipgloss.Color("#a6adc8")
	if focused {
		borderColor = lipgloss.Color("#89b4fa")
		titleColor = lipgloss.Color("#89b4fa")
	}

	title := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(focused).
		Render("COMMITS [3]")

	inner := m.commits.View()

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2).
		Render(title + "\n" + inner)
}

func (m Model) renderDiffPanel(width, height int) string {
	focused := m.focused == PanelDiff
	borderColor := lipgloss.Color("#585b70")
	titleColor := lipgloss.Color("#a6adc8")
	if focused {
		borderColor = lipgloss.Color("#89b4fa")
		titleColor = lipgloss.Color("#89b4fa")
	}

	stagedLabel := ""
	if m.diffStaged {
		stagedLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Render(" [staged]")
	}

	title := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(focused).
		Render("DIFF") + stagedLabel

	inner := m.diff.View()

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2).
		Render(title + "\n" + inner)
}

func (m Model) renderFooter() string {
	hints := []widgets.KeyHint{
		{Key: "tab", Desc: "panel"},
		{Key: "space", Desc: "stage"},
		{Key: "c", Desc: "commit"},
		{Key: "P/p", Desc: "push/pull"},
		{Key: "r", Desc: "refresh"},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	}

	if m.mode == ModeCommit {
		hints = []widgets.KeyHint{
			{Key: "ctrl+s", Desc: "commit"},
			{Key: "esc", Desc: "cancel"},
		}
	}

	sb := widgets.NewStatusBar(m.width)
	sb.Hints = hints
	sb.Extra = m.statusLine()

	return sb.View()
}

func (m Model) statusLine() string {
	if m.statusErr {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render(m.statusMsg)
	}
	if m.statusMsg != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Render(m.statusMsg)
	}
	return m.files.Stats()
}

// --- Panel focus rotation ---

func (m *Model) focusNext() {
	switch m.focused {
	case PanelFiles:
		m.focused = PanelBranches
	case PanelBranches:
		m.focused = PanelCommits
	case PanelCommits:
		m.focused = PanelDiff
	case PanelDiff:
		m.focused = PanelFiles
	}
}

func (m *Model) focusPrev() {
	switch m.focused {
	case PanelFiles:
		m.focused = PanelDiff
	case PanelBranches:
		m.focused = PanelFiles
	case PanelCommits:
		m.focused = PanelBranches
	case PanelDiff:
		m.focused = PanelCommits
	}
}

// rebuildPanels recreates panel models with updated dimensions.
func (m *Model) rebuildPanels() {
	layout := ComputeLayout(m.width, m.height)
	panelH := layout.LeftHeight
	leftW := layout.LeftWidth
	rightW := layout.RightWidth

	tabH := panelH / 3
	if tabH < 4 {
		tabH = 4
	}

	m.files = panels.NewFilesModel(leftW, tabH)
	m.branches = panels.NewBranchModel(leftW, tabH)
	m.commits = panels.NewCommitModel(leftW, panelH-2*tabH)
	m.diff = panels.NewDiffModel(rightW, panelH)
	m.commitMsg = panels.NewCommitMsgModel(leftW, tabH)
	m.statusBar = widgets.NewStatusBar(m.width)
}
