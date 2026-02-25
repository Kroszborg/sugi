package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/ai"
	"github.com/Kroszborg/sugi/internal/config"
	"github.com/Kroszborg/sugi/internal/forge"
	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui/panels"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Messages ---

type GitStatusMsg struct {
	Files []git.FileStatus
	Err   error
}
type GitBranchesMsg struct {
	Branches []git.Branch
	Err      error
}
type GitCommitsMsg struct {
	Commits []git.Commit
	Err     error
}
type GitGraphMsg struct {
	Lines []string
	Err   error
}
type GitDiffMsg struct {
	Hunks  []git.DiffHunk
	Staged bool
	Err    error
}
type GitStashMsg struct {
	Stashes []git.StashEntry
	Err     error
}
type GitStashDiffMsg struct {
	FileDiffs []git.FileDiff
	Err       error
}
type GitBlameMsg struct {
	Lines []git.BlameLine
	Err   error
}
type GitReflogMsg struct {
	Entries []git.ReflogEntry
	Err     error
}
type GitOperationMsg struct {
	Op  string
	Err error
}
type GitPRsMsg struct {
	PRs []forge.PullRequest
	Err error
}
type StatusMsg struct {
	Text  string
	IsErr bool
}

// AI streaming messages
type AIChunkMsg struct{ Text string }
type AIDoneMsg struct{}
type AIErrorMsg struct{ Err error }

// AISummaryMsg is emitted when the diff summary generation completes.
type AISummaryMsg struct{ Text string }

// SettingsSavedMsg is emitted after settings are written to disk.
type SettingsSavedMsg struct{ Cfg config.Config }

func waitForAI(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return AIDoneMsg{}
		}
		return msg
	}
}

// --- Mode ---

type Mode int

const (
	ModeNormal Mode = iota
	ModeCommit
	ModeHelp
	ModeNewBranch
	ModeStash
	ModeBlame
	ModeReflog
	ModeSearch
	ModePR
	ModeAIGenerating
	ModePalette
	ModeAISummary
	ModeSettings
)

// Extended PanelIDs (beyond the core 0-4 in layout.go)
const (
	PanelStash  PanelID = 10
	PanelBlame  PanelID = 11
	PanelReflog PanelID = 12
	PanelPR     PanelID = 13
)

// --- Model ---

type Model struct {
	repo   *git.Client
	cfg    config.Config
	forge  forge.ForgeClient
	aiGen  *ai.Generator
	keymap KeyMap

	files     panels.FilesModel
	branches  panels.BranchModel
	commits   panels.CommitModel
	diff      panels.DiffModel
	commitMsg panels.CommitMsgModel
	stash     panels.StashModel
	blame     panels.BlameModel
	reflog    panels.ReflogModel
	pr        panels.PRModel
	palette   panels.PaletteModel
	settings  panels.SettingsModel

	help        widgets.HelpOverlay
	searchInput textinput.Model
	searchPanel PanelID

	focused     PanelID
	prevFocused PanelID
	mode        Mode

	currentFilePath string
	diffStaged      bool

	width  int
	height int

	statusMsg string
	statusErr bool
	loading   bool
	loadingMsg string

	aiGenerating bool
	aiBuffer     string
	aiChan       chan tea.Msg

	aiSummary string // AI diff summary text

	theme Theme
}

// New creates the root model. forgeClient and aiGen may be nil.
func New(repo *git.Client, cfg config.Config, forgeClient forge.ForgeClient, aiGen *ai.Generator) Model {
	m := Model{
		repo:   repo,
		cfg:    cfg,
		forge:  forgeClient,
		aiGen:  aiGen,
		keymap: DefaultKeyMap(),
		theme:  DefaultTheme,
		width:  80,
		height: 24,
	}
	si := textinput.New()
	si.Placeholder = "search…"
	si.CharLimit = 100
	m.searchInput = si
	m.rebuildPanels()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildPanels()
		return m, tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits())

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)

	case GitStatusMsg:
		m.loading = false
		if msg.Err != nil {
			m.setStatus("Error: "+msg.Err.Error(), true)
		} else {
			m.files.SetFiles(msg.Files)
		}

	case GitBranchesMsg:
		if msg.Err == nil {
			m.branches.SetBranches(msg.Branches)
		}

	case GitCommitsMsg:
		if msg.Err == nil {
			m.commits.SetCommits(msg.Commits)
		}

	case GitGraphMsg:
		if msg.Err == nil {
			m.commits.SetGraphLines(msg.Lines)
		}

	case GitDiffMsg:
		if msg.Err == nil {
			m.diff.SetHunks(msg.Hunks)
		}

	case GitStashMsg:
		if msg.Err == nil {
			m.stash.SetStashes(msg.Stashes)
		}

	case GitStashDiffMsg:
		if msg.Err == nil {
			m.stash.SetDiff(msg.FileDiffs)
		}

	case GitBlameMsg:
		if msg.Err == nil {
			m.blame.SetBlame(msg.Lines)
		}

	case GitReflogMsg:
		if msg.Err == nil {
			m.reflog.SetEntries(msg.Entries)
		}

	case GitPRsMsg:
		if msg.Err == nil {
			m.pr.SetPRs(msg.PRs)
		} else {
			m.setStatus("PRs: "+msg.Err.Error(), true)
		}

	case GitOperationMsg:
		m.loading = false
		if msg.Err != nil {
			m.setStatus(msg.Op+" failed: "+msg.Err.Error(), true)
		} else {
			m.setStatus(msg.Op+" ✓", false)
		}
		return m, tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits())

	case AIChunkMsg:
		m.aiBuffer += msg.Text
		m.commitMsg.SetValue(m.aiBuffer)
		m.commitMsg.AIGenerating = true
		if m.aiChan != nil {
			return m, waitForAI(m.aiChan)
		}

	case AIDoneMsg:
		m.aiGenerating = false
		m.mode = ModeCommit
		m.aiChan = nil
		m.commitMsg.AIGenerating = false
		m.commitMsg.Focus()
		if m.aiBuffer == "" {
			m.setStatus("AI error: no response — set groq_api_key in settings (O) or ~/.config/sugi/config.json", true)
		} else {
			// Clean up the AI output: deduplicate repeated paragraphs, strip preamble.
			cleaned := cleanCommitMsg(m.aiBuffer)
			m.aiBuffer = cleaned
			m.commitMsg.SetValue(cleaned)
			m.setStatus("✓ AI done — review and ctrl+s to commit", false)
		}

	case AIErrorMsg:
		m.aiGenerating = false
		m.mode = ModeCommit
		m.aiChan = nil
		m.commitMsg.Focus()
		m.setStatus("AI error: "+msg.Err.Error(), true)

	case StatusMsg:
		if msg.Text != "" { // ignore empty no-op messages
			m.setStatus(msg.Text, msg.IsErr)
		}

	case AISummaryMsg:
		if msg.Text == "" {
			m.aiSummary = "AI returned empty summary — check your AI backend"
		} else {
			m.aiSummary = msg.Text
		}
		m.mode = ModeAISummary

	case panels.PaletteSelectMsg:
		m.mode = ModeNormal
		return m.executePaletteAction(msg.ID)

	case SettingsSavedMsg:
		m.aiGen = ai.NewGenerator(msg.Cfg.GroqAPIKey, msg.Cfg.GroqModel)
		m.setStatus("Settings saved ✓", false)
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}
	base := lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(), m.renderBody(), m.renderFooter())
	switch m.mode {
	case ModeHelp:
		return m.renderCentered(m.help.View())
	case ModePalette:
		return m.renderCentered(m.palette.View())
	case ModeAISummary:
		return m.renderCentered(m.renderAISummary())
	case ModeSettings:
		return m.renderCentered(m.settings.View())
	}
	return base
}

// ─── Key handling ────────────────────────────────────────────────────────────

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help overlay captures all input.
	if m.mode == ModeHelp {
		switch msg.String() {
		case "?", "esc":
			m.mode = ModeNormal
		case "j", "down":
			m.help.ScrollDown()
		case "k", "up":
			m.help.ScrollUp()
		}
		return m, nil
	}
	// Palette captures all input.
	if m.mode == ModePalette {
		return m.handlePaletteKey(msg)
	}
	// AI summary overlay — esc closes it.
	if m.mode == ModeAISummary {
		if key.Matches(msg, m.keymap.Escape) {
			m.mode = ModeNormal
			m.aiSummary = ""
		}
		return m, nil
	}
	// Settings overlay.
	if m.mode == ModeSettings {
		return m.handleSettingsKey(msg)
	}
	if m.mode == ModeSearch {
		return m.handleSearchKey(msg)
	}
	if m.mode == ModeCommit || m.mode == ModeAIGenerating {
		return m.handleCommitKey(msg)
	}
	if m.mode == ModeNewBranch {
		return m.handleNewBranchKey(msg)
	}

	switch {
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keymap.Help), msg.String() == "?":
		if m.mode == ModeHelp {
			m.mode = ModeNormal
		} else {
			m.mode = ModeHelp
		}
		return m, nil
	case key.Matches(msg, m.keymap.Escape):
		return m.handleEscape()
	case key.Matches(msg, m.keymap.Refresh):
		m.loading = true
		m.loadingMsg = "Refreshing…"
		return m, tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits())
	case key.Matches(msg, m.keymap.Search):
		m.mode = ModeSearch
		m.searchPanel = m.focused
		m.searchInput.Reset()
		m.searchInput.Focus()
		return m, nil
	case key.Matches(msg, m.keymap.FocusFiles):
		m.focused = PanelFiles
		return m, m.loadDiffForCursor()
	case key.Matches(msg, m.keymap.FocusBranches):
		m.focused = PanelBranches
		return m, nil
	case key.Matches(msg, m.keymap.FocusCommits):
		m.focused = PanelCommits
		return m, m.loadDiffForCursor()
	case key.Matches(msg, m.keymap.FocusDiff):
		m.focused = PanelDiff
		return m, nil
	case key.Matches(msg, m.keymap.NextPanel):
		m.cyclePanel(1)
		return m, m.loadDiffForCursor()
	case key.Matches(msg, m.keymap.PrevPanel):
		m.cyclePanel(-1)
		return m, m.loadDiffForCursor()
	// P opens PRs only when not in Files/Branches (where P means Push).
	case msg.String() == "P" && m.focused != PanelFiles && m.focused != PanelBranches:
		return m.openPRPanel()
	case key.Matches(msg, m.keymap.Palette):
		m.mode = ModePalette
		m.palette.Open()
		return m, nil
	case key.Matches(msg, m.keymap.Settings):
		m.settings.LoadConfig(m.cfg)
		m.mode = ModeSettings
		return m, nil
	}

	switch m.focused {
	case PanelStash:
		return m.handleStashKey(msg)
	case PanelBlame:
		return m.handleBlameKey(msg)
	case PanelReflog:
		return m.handleReflogKey(msg)
	case PanelPR:
		return m.handlePRKey(msg)
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

func (m Model) handleEscape() (tea.Model, tea.Cmd) {
	switch m.focused {
	case PanelStash, PanelBlame, PanelReflog, PanelPR:
		m.mode = ModeNormal
		m.focused = m.prevFocused
		return m, nil
	}
	m.mode = ModeNormal
	return m, nil
}

func (m Model) handleFilesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Down):
		m.files.MoveDown()
		return m, m.loadDiffForCursor()
	case key.Matches(msg, m.keymap.Up):
		m.files.MoveUp()
		return m, m.loadDiffForCursor()
	case key.Matches(msg, m.keymap.Stage):
		return m, m.toggleStage()
	case key.Matches(msg, m.keymap.StageAll):
		return m, m.stageAll()
	case key.Matches(msg, m.keymap.Discard):
		return m, m.discardFile()
	case key.Matches(msg, m.keymap.Commit):
		m.mode = ModeCommit
		m.commitMsg.Focus()
		return m, nil
	case key.Matches(msg, m.keymap.Push):
		return m, m.push()
	case key.Matches(msg, m.keymap.Pull):
		return m, m.pull()
	case key.Matches(msg, m.keymap.Fetch):
		return m, m.fetch()
	case key.Matches(msg, m.keymap.ToggleDiffStaged):
		m.diffStaged = !m.diffStaged
		return m, m.loadDiffForCursor()
	case msg.String() == "z":
		m.prevFocused = m.focused
		m.focused = PanelStash
		m.mode = ModeStash
		return m, m.loadStashes()
	case msg.String() == "P":
		return m.openPRPanel()
	}
	return m, nil
}

func (m Model) handleBranchesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Down):
		m.branches.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.branches.MoveUp()
	case key.Matches(msg, m.keymap.Checkout):
		b := m.branches.CurrentBranch()
		if b != nil && !b.IsCurrent {
			return m, m.checkout(b.Name)
		}
	case key.Matches(msg, m.keymap.NewBranch):
		m.mode = ModeNewBranch
		m.branches.ShowNewBranchModal()
		return m, nil
	case key.Matches(msg, m.keymap.Delete):
		b := m.branches.CurrentBranch()
		if b != nil && !b.IsCurrent {
			return m, m.deleteBranch(b.Name)
		}
	case key.Matches(msg, m.keymap.Push):
		return m, m.push()
	case key.Matches(msg, m.keymap.Pull):
		return m, m.pull()
	case msg.String() == "P":
		return m.openPRPanel()
	}
	return m, nil
}

func (m Model) handleCommitsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Down):
		m.commits.MoveDown()
		return m, m.loadCommitDiff()
	case key.Matches(msg, m.keymap.Up):
		m.commits.MoveUp()
		return m, m.loadCommitDiff()
	case key.Matches(msg, m.keymap.PageDown):
		m.commits.PageDown()
		return m, m.loadCommitDiff()
	case key.Matches(msg, m.keymap.PageUp):
		m.commits.PageUp()
		return m, m.loadCommitDiff()
	case msg.String() == "g":
		m.commits.ToggleGraph()
		if m.commits.ShowGraph {
			return m, m.loadGraph()
		}
	case msg.String() == "b":
		f := m.files.CurrentFile()
		if f != nil {
			m.prevFocused = m.focused
			m.focused = PanelBlame
			m.mode = ModeBlame
			path := f.Path
			return m, func() tea.Msg {
				lines, err := m.repo.Blame(path)
				return GitBlameMsg{Lines: lines, Err: err}
			}
		}
	case msg.String() == "R":
		m.prevFocused = m.focused
		m.focused = PanelReflog
		m.mode = ModeReflog
		return m, m.loadReflog()
	}
	return m, nil
}

func (m Model) handleDiffKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Down):
		m.diff.ScrollDown()
	case key.Matches(msg, m.keymap.Up):
		m.diff.ScrollUp()
	case key.Matches(msg, m.keymap.PageDown):
		m.diff.PageDown()
	case key.Matches(msg, m.keymap.PageUp):
		m.diff.PageUp()
	case msg.String() == "]":
		m.diff.NextHunk()
	case msg.String() == "[":
		m.diff.PrevHunk()
	case key.Matches(msg, m.keymap.Stage):
		return m, m.stageHunk(false)
	case key.Matches(msg, m.keymap.Unstage):
		return m, m.stageHunk(true)
	case msg.String() == "A":
		m.mode = ModeAISummary
		m.aiSummary = "⟳ Generating AI summary…"
		return m, m.summarizeDiff()
	}
	return m, nil
}

func (m Model) handleStashKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.stash.MoveDown()
		return m, m.loadStashDiff()
	case key.Matches(msg, m.keymap.Up):
		m.stash.MoveUp()
		return m, m.loadStashDiff()
	case key.Matches(msg, m.keymap.Checkout):
		s := m.stash.CurrentStash()
		if s != nil {
			ref := s.Ref
			return m, func() tea.Msg {
				return GitOperationMsg{Op: "Stash apply", Err: m.repo.StashApply(ref)}
			}
		}
	case key.Matches(msg, m.keymap.Delete):
		s := m.stash.CurrentStash()
		if s != nil {
			ref := s.Ref
			return m, func() tea.Msg {
				return GitOperationMsg{Op: "Stash drop", Err: m.repo.StashDrop(ref)}
			}
		}
	case msg.String() == "p":
		return m, func() tea.Msg {
			return GitOperationMsg{Op: "Stash pop", Err: m.repo.StashPop()}
		}
	}
	return m, nil
}

func (m Model) handleBlameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.blame.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.blame.MoveUp()
	case key.Matches(msg, m.keymap.PageDown):
		m.blame.PageDown()
	case key.Matches(msg, m.keymap.PageUp):
		m.blame.PageUp()
	}
	return m, nil
}

func (m Model) handleReflogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.reflog.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.reflog.MoveUp()
	case msg.String() == "u":
		e := m.reflog.CurrentEntry()
		if e != nil {
			ref := e.Ref
			return m, func() tea.Msg {
				return GitOperationMsg{Op: "Undo to " + ref, Err: m.repo.ReflogUndo(ref)}
			}
		}
	}
	return m, nil
}

func (m Model) handlePRKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.pr.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.pr.MoveUp()
	case key.Matches(msg, m.keymap.Checkout):
		m.pr.ToggleDetail()
	case msg.String() == "m":
		pr := m.pr.CurrentPR()
		if pr != nil && m.forge != nil {
			num, fc := pr.Number, m.forge
			return m, func() tea.Msg {
				return GitOperationMsg{Op: fmt.Sprintf("Merge PR #%d", num), Err: fc.MergePR(num, "merge")}
			}
		}
	case msg.String() == "x":
		pr := m.pr.CurrentPR()
		if pr != nil && m.forge != nil {
			num, fc := pr.Number, m.forge
			return m, func() tea.Msg {
				return GitOperationMsg{Op: fmt.Sprintf("Close PR #%d", num), Err: fc.ClosePR(num)}
			}
		}
	case msg.String() == "r":
		return m, m.loadPRs()
	}
	return m, nil
}

func (m Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settings.IsEditing() {
		switch msg.String() {
		case "esc":
			m.settings.CancelEdit()
			return m, nil
		case "enter", "ctrl+s":
			// Confirm the edit and auto-save immediately.
			return m.saveSettings()
		default:
			var cmd tea.Cmd
			m.settings, cmd = m.settings.UpdateInput(msg)
			return m, cmd
		}
	}
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil
	case "up", "k":
		m.settings.MoveUp()
	case "down", "j":
		m.settings.MoveDown()
	case "enter", "e":
		m.settings.StartEdit()
	case " ":
		m.settings.Toggle()
	case "ctrl+s":
		return m.saveSettings()
	}
	return m, nil
}

func (m Model) saveSettings() (tea.Model, tea.Cmd) {
	// Always confirm any active text edit before building config.
	m.settings.ConfirmEdit()
	newCfg := m.settings.BuildConfig(m.cfg)
	m.cfg = newCfg
	// Update AI generator immediately so ctrl+g works right after saving.
	m.aiGen = ai.NewGenerator(newCfg.GroqAPIKey, newCfg.GroqModel)
	m.mode = ModeNormal
	keyHint := ""
	if newCfg.GroqAPIKey != "" {
		keyHint = "  Groq key set ✓"
	} else {
		keyHint = "  no Groq key — AI won't work"
	}
	m.setStatus("Settings saved"+keyHint, newCfg.GroqAPIKey == "")
	return m, func() tea.Msg {
		if err := config.Save(newCfg); err != nil {
			return StatusMsg{Text: "Error writing config file: " + err.Error(), IsErr: true}
		}
		return StatusMsg{} // no-op, config already shown above
	}
}

func (m Model) handleCommitKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == ModeAIGenerating {
		if msg.String() == "esc" {
			m.aiGenerating = false
			m.mode = ModeCommit
			m.aiChan = nil
			m.commitMsg.AIGenerating = false
			m.setStatus("AI cancelled", false)
		}
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.commitMsg.Blur()
		m.commitMsg.Reset()
		m.aiBuffer = ""
		return m, nil
	case "tab":
		m.commitMsg.NextField()
		return m, nil
	case "ctrl+g", "alt+g":
		ch := make(chan tea.Msg, 64)
		m.aiChan = ch
		m.aiGenerating = true
		m.mode = ModeAIGenerating
		m.aiBuffer = ""
		m.commitMsg.SetValue("")
		aiGen, repo := m.aiGen, m.repo
		return m, generateCommitMsgCmd(aiGen, repo, ch)
	case "ctrl+s":
		if strings.TrimSpace(m.commitMsg.Subject()) == "" {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Subject line cannot be empty", IsErr: true}
			}
		}
		m.mode = ModeNormal
		m.commitMsg.Blur()
		msgText := m.commitMsg.Value()
		m.aiBuffer = ""
		m.commitMsg.Reset()
		return m, m.doCommit(msgText)
	}
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
	return m, m.branches.UpdateModalInput(msg)
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.searchInput.Blur()
		m.commits.SetFilter("") // cancel: clear filter
		return m, nil
	case "enter":
		m.mode = ModeNormal
		m.searchInput.Blur()
		return m, nil // keep filter applied
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	if m.searchPanel == PanelCommits {
		m.commits.SetFilter(m.searchInput.Value())
	}
	return m, cmd
}

// ─── Mouse ───────────────────────────────────────────────────────────────────

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		switch m.focused {
		case PanelFiles:
			m.files.MoveUp()
			return m, m.loadDiffForCursor()
		case PanelBranches:
			m.branches.MoveUp()
		case PanelCommits:
			m.commits.MoveUp()
			return m, m.loadCommitDiff()
		case PanelDiff:
			m.diff.ScrollUp()
		case PanelPR:
			m.pr.MoveUp()
		}
	case tea.MouseButtonWheelDown:
		switch m.focused {
		case PanelFiles:
			m.files.MoveDown()
			return m, m.loadDiffForCursor()
		case PanelBranches:
			m.branches.MoveDown()
		case PanelCommits:
			m.commits.MoveDown()
			return m, m.loadCommitDiff()
		case PanelDiff:
			m.diff.ScrollDown()
		case PanelPR:
			m.pr.MoveDown()
		}
	}
	return m, nil
}

// ─── Git commands as tea.Cmd ─────────────────────────────────────────────────

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
		commits, err := m.repo.Log(200)
		return GitCommitsMsg{Commits: commits, Err: err}
	}
}

func (m Model) loadGraph() tea.Cmd {
	return func() tea.Msg {
		lines, err := m.repo.LogGraph(100)
		return GitGraphMsg{Lines: lines, Err: err}
	}
}

func (m Model) loadDiffForCursor() tea.Cmd {
	switch m.focused {
	case PanelFiles:
		f := m.files.CurrentFile()
		if f == nil {
			return func() tea.Msg { return GitDiffMsg{} }
		}
		m.currentFilePath = f.Path
		path, staged := f.Path, m.diffStaged
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
		var all []git.DiffHunk
		for _, fd := range fds {
			all = append(all, fd.Hunks...)
		}
		return GitDiffMsg{Hunks: all}
	}
}

func (m Model) loadStashes() tea.Cmd {
	return func() tea.Msg {
		s, err := m.repo.Stashes()
		return GitStashMsg{Stashes: s, Err: err}
	}
}

func (m Model) loadStashDiff() tea.Cmd {
	s := m.stash.CurrentStash()
	if s == nil {
		return nil
	}
	ref := s.Ref
	return func() tea.Msg {
		fds, err := m.repo.StashShow(ref)
		return GitStashDiffMsg{FileDiffs: fds, Err: err}
	}
}

func (m Model) loadReflog() tea.Cmd {
	return func() tea.Msg {
		e, err := m.repo.Reflog(100)
		return GitReflogMsg{Entries: e, Err: err}
	}
}

func (m Model) loadPRs() tea.Cmd {
	if m.forge == nil {
		return func() tea.Msg {
			return GitPRsMsg{Err: fmt.Errorf("no forge detected — set GITHUB_TOKEN or GITLAB_TOKEN")}
		}
	}
	fc := m.forge
	return func() tea.Msg {
		prs, err := fc.ListPRs("open")
		return GitPRsMsg{PRs: prs, Err: err}
	}
}

func (m Model) toggleStage() tea.Cmd {
	f := m.files.CurrentFile()
	if f == nil {
		return nil
	}
	path, isStaged := f.Path, f.IsStaged()
	return func() tea.Msg {
		op := "Stage"
		var err error
		if isStaged {
			op = "Unstage"
			err = m.repo.Unstage(path)
		} else {
			err = m.repo.Stage(path)
		}
		return GitOperationMsg{Op: op, Err: err}
	}
}

func (m Model) stageAll() tea.Cmd {
	return func() tea.Msg {
		return GitOperationMsg{Op: "Stage all", Err: m.repo.StageAll()}
	}
}

func (m Model) discardFile() tea.Cmd {
	f := m.files.CurrentFile()
	if f == nil {
		return nil
	}
	path := f.Path
	return func() tea.Msg {
		return GitOperationMsg{Op: "Discard " + path, Err: m.repo.DiscardFile(path)}
	}
}

func (m Model) stageHunk(reverse bool) tea.Cmd {
	if m.currentFilePath == "" {
		return nil
	}
	patch := m.diff.BuildHunkPatch(m.currentFilePath, m.diff.CurrentHunkIndex(), reverse)
	if patch == "" {
		return nil
	}
	op := "Stage hunk"
	if reverse {
		op = "Unstage hunk"
	}
	return func() tea.Msg {
		var err error
		if reverse {
			err = m.repo.UnstageHunk(patch)
		} else {
			err = m.repo.StageHunk(patch)
		}
		return GitOperationMsg{Op: op, Err: err}
	}
}

func (m Model) doCommit(message string) tea.Cmd {
	return func() tea.Msg {
		return GitOperationMsg{Op: "Commit", Err: m.repo.Commit(message)}
	}
}

func (m Model) push() tea.Cmd {
	return func() tea.Msg { return GitOperationMsg{Op: "Push", Err: m.repo.Push()} }
}

func (m Model) pull() tea.Cmd {
	return func() tea.Msg { return GitOperationMsg{Op: "Pull", Err: m.repo.Pull()} }
}

func (m Model) fetch() tea.Cmd {
	return func() tea.Msg { return GitOperationMsg{Op: "Fetch", Err: m.repo.Fetch()} }
}

func (m Model) checkout(name string) tea.Cmd {
	return func() tea.Msg {
		return GitOperationMsg{Op: "Checkout " + name, Err: m.repo.Checkout(name)}
	}
}

func (m Model) createBranch(name string) tea.Cmd {
	return func() tea.Msg {
		return GitOperationMsg{Op: "New branch " + name, Err: m.repo.CreateBranch(name)}
	}
}

func (m Model) deleteBranch(name string) tea.Cmd {
	return func() tea.Msg {
		return GitOperationMsg{Op: "Delete branch " + name, Err: m.repo.DeleteBranch(name)}
	}
}

// generateCommitMsgCmd builds the staged diff, starts an AI goroutine, and
// blocks on the first response so Bubbletea gets a real message to dispatch.
// Errors are sent through the channel so Update can display them properly.
func generateCommitMsgCmd(aiGen *ai.Generator, repo *git.Client, ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		// Build staged diff; fall back to unstaged if nothing is staged.
		fds, _ := repo.DiffAll(true)
		if len(fds) == 0 {
			fds, _ = repo.DiffAll(false)
		}
		var sb strings.Builder
		for _, fd := range fds {
			sb.WriteString("--- " + fd.OldPath + "\n+++ " + fd.NewPath + "\n")
			for _, h := range fd.Hunks {
				sb.WriteString(h.Header + "\n")
				for _, dl := range h.Lines {
					switch dl.Type {
					case git.DiffAdded:
						sb.WriteString("+" + dl.Content + "\n")
					case git.DiffRemoved:
						sb.WriteString("-" + dl.Content + "\n")
					case git.DiffContext:
						sb.WriteString(" " + dl.Content + "\n")
					}
				}
			}
		}

		go func() {
			defer close(ch)
			err := aiGen.Generate(context.Background(), ai.CommitMsgPrompt(sb.String()), func(s string) {
				ch <- AIChunkMsg{Text: s}
			})
			if err != nil {
				ch <- AIErrorMsg{Err: err}
			}
		}()

		// Block on first message so Bubbletea gets something to dispatch.
		msg, ok := <-ch
		if !ok {
			return AIDoneMsg{}
		}
		return msg
	}
}

// ─── View rendering ───────────────────────────────────────────────────────────

func (m Model) renderCentered(overlay string) string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(ColorBase)),
	)
}

func (m Model) renderAISummary() string {
	w := m.width - 6
	if w < 40 {
		w = 40
	}
	h := m.height - 6
	if h < 5 {
		h = 5
	}

	title := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTeal)).Bold(true).Render("  ✦ AI Diff Summary")

	body := m.aiSummary
	// word-wrap at width
	lines := strings.Split(body, "\n")
	var wrapped []string
	for _, line := range lines {
		for len(line) > w-4 {
			wrapped = append(wrapped, line[:w-4])
			line = line[w-4:]
		}
		wrapped = append(wrapped, line)
	}
	// trim to fit height
	maxLines := h - 4
	if maxLines < 1 {
		maxLines = 1
	}
	if len(wrapped) > maxLines {
		wrapped = wrapped[:maxLines]
		wrapped = append(wrapped, lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render("… (truncated)"))
	}

	content := strings.Join(wrapped, "\n")
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render("\n[esc] close")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#94e2d5")).
		Width(w).Height(h).
		Padding(0, 1).
		Render(title + "\n\n" + content + hint)
}

func (m Model) renderHeader() string {
	// Left: repo + branch
	repoStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#181825")).
		Foreground(lipgloss.Color("#89b4fa")).Bold(true)
	branchStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#181825")).
		Foreground(lipgloss.Color("#a6e3a1"))
	mutedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#181825")).
		Foreground(lipgloss.Color("#45475a"))

	repoName := repoStyle.Render("  " + m.repo.RepoName())
	branch, _ := m.repo.CurrentBranch()
	sep := mutedStyle.Render("  ⎇ ")
	branchStr := branchStyle.Render(branch)

	extra := ""
	if m.loading {
		extra = lipgloss.NewStyle().Background(lipgloss.Color("#181825")).
			Foreground(lipgloss.Color("#f9e2af")).Render("   ⟳ " + m.loadingMsg)
	}
	if m.mode == ModeSearch {
		extra += lipgloss.NewStyle().Background(lipgloss.Color("#181825")).
			Foreground(lipgloss.Color("#cba6f7")).Render("   / " + m.searchInput.Value())
	}
	if m.aiGenerating {
		extra += lipgloss.NewStyle().Background(lipgloss.Color("#181825")).
			Foreground(lipgloss.Color("#94e2d5")).Render("   ✦ AI…")
	}

	forgeStr := ""
	if m.forge != nil {
		info := m.forge.ForgeInfo()
		forgeStr = mutedStyle.Render("   " + info.Type.String())
	}

	right := lipgloss.NewStyle().
		Background(lipgloss.Color("#181825")).
		Foreground(lipgloss.Color("#45475a")).
		Render("sugi  ")

	left := repoName + sep + branchStr + forgeStr + extra
	pad := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	bg := lipgloss.NewStyle().Background(lipgloss.Color("#181825"))
	return bg.Width(m.width).Render(left + strings.Repeat(" ", pad) + right)
}

func (m Model) renderBody() string {
	layout := ComputeLayout(m.width, m.height)

	switch m.focused {
	case PanelStash:
		return m.renderOverlay("STASH  enter:apply  p:pop  D:drop  esc:close", m.stash.View(), layout)
	case PanelBlame:
		return m.renderOverlay("BLAME  ↑↓:navigate  esc:close  "+m.blame.StatusInfo(), m.blame.View(), layout)
	case PanelReflog:
		return m.renderOverlay("REFLOG  u:undo to here  esc:close", m.reflog.View(), layout)
	case PanelPR:
		hint := "PULL REQUESTS  enter:detail  m:merge  x:close  r:refresh  esc:close"
		if m.forge == nil {
			hint = "PULL REQUESTS  — set GITHUB_TOKEN or GITLAB_TOKEN  esc:close"
		}
		return m.renderOverlay(hint, m.pr.View(), layout)
	}

	if m.mode == ModeCommit || m.mode == ModeAIGenerating {
		return m.renderCommitMode(layout)
	}

	panelH, leftW, rightW := layout.LeftHeight, layout.LeftWidth, layout.RightWidth
	tabH := panelH / 3
	if tabH < 5 {
		tabH = 5
	}

	left := lipgloss.JoinVertical(lipgloss.Left,
		m.renderFilesPanel(leftW, tabH),
		m.renderBranchesPanel(leftW, tabH),
		m.renderCommitsPanel(leftW, panelH-2*tabH),
	)

	if layout.IsVeryNarrow || rightW == 0 {
		return left
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, m.renderDiffPanel(rightW, panelH))
}

func (m Model) renderCommitMode(layout Layout) string {
	panelH, leftW, rightW := layout.LeftHeight, layout.LeftWidth, layout.RightWidth

	// Border and title color
	isGenerating := m.mode == ModeAIGenerating
	borderClr := lipgloss.Color("#89dceb") // sky
	if isGenerating {
		borderClr = lipgloss.Color("#94e2d5") // teal while generating
	}

	// Title row
	titleText := lipgloss.NewStyle().Foreground(borderClr).Bold(true).Render(" COMMIT")
	var titleSuffix string
	if isGenerating {
		titleSuffix = lipgloss.NewStyle().Foreground(lipgloss.Color("#94e2d5")).
			Render("   ✦ AI generating…")
	} else {
		titleSuffix = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).
			Render("   conventional commits")
	}
	title := titleText + titleSuffix

	left := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderClr).
		Width(leftW - 2).Height(panelH - 2).
		Render(title + "\n\n" + m.commitMsg.View())
	if layout.IsVeryNarrow || rightW == 0 {
		return left
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, m.renderDiffPanel(rightW, panelH))
}

func (m Model) renderOverlay(title, content string, layout Layout) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorSky)).
		Width(m.width - 2).Height(layout.LeftHeight - 2).
		Render(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSky)).Bold(true).Render(" "+title) + "\n" + content)
}

func (m Model) renderFilesPanel(width, height int) string {
	focused := m.focused == PanelFiles && m.mode == ModeNormal
	bc, tc := panelColors(focused)
	stats := m.files.Stats()
	statsStr := ""
	if stats != "" {
		statsStr = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render("  " + stats)
	}
	title := lipgloss.NewStyle().Foreground(tc).Bold(focused).Render(" FILES") + statsStr +
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOverlay)).Render("  [1]")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(bc).
		Width(width - 2).Height(height - 2).
		Render(title + "\n" + m.files.View())
}

func (m Model) renderBranchesPanel(width, height int) string {
	focused := m.focused == PanelBranches && m.mode == ModeNormal
	bc, tc := panelColors(focused)
	title := lipgloss.NewStyle().Foreground(tc).Bold(focused).Render(" BRANCHES") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOverlay)).Render("  [2]")
	inner := m.branches.View()
	if m.mode == ModeNewBranch {
		inner += "\n" + m.branches.ModalView()
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(bc).
		Width(width - 2).Height(height - 2).
		Render(title + "\n" + inner)
}

func (m Model) renderCommitsPanel(width, height int) string {
	focused := m.focused == PanelCommits && m.mode == ModeNormal
	bc, tc := panelColors(focused)
	extra := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOverlay)).Render("  [3]")
	if m.commits.ShowGraph {
		extra = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTeal)).Render("  graph") +
			lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOverlay)).Render("  [3]")
	}
	title := lipgloss.NewStyle().Foreground(tc).Bold(focused).Render(" COMMITS") + extra
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(bc).
		Width(width - 2).Height(height - 2).
		Render(title + "\n" + m.commits.View())
}

func (m Model) renderDiffPanel(width, height int) string {
	focused := m.focused == PanelDiff
	bc, tc := panelColors(focused)
	extra := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOverlay)).Render("  [4]")
	if m.diffStaged {
		extra = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Render("  staged") + extra
	}
	if m.diff.HunkCount() > 0 {
		extra += lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).
			Render("  " + m.diff.ScrollInfo())
	}
	title := lipgloss.NewStyle().Foreground(tc).Bold(focused).Render(" DIFF") + extra
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(bc).
		Width(width - 2).Height(height - 2).
		Render(title + "\n" + m.diff.View())
}

func (m Model) renderFooter() string {
	var hints []widgets.KeyHint

	switch m.mode {
	case ModeCommit, ModeAIGenerating:
		hints = []widgets.KeyHint{
			{Key: "tab", Desc: "next field"},
			{Key: "ctrl+g/alt+g", Desc: "AI generate"},
			{Key: "ctrl+s", Desc: "commit"},
			{Key: "esc", Desc: "cancel"},
		}
	case ModeSearch:
		hints = []widgets.KeyHint{
			{Key: "type", Desc: "filter"},
			{Key: "enter", Desc: "apply"},
			{Key: "esc", Desc: "cancel"},
		}
	case ModePalette:
		hints = []widgets.KeyHint{
			{Key: "type", Desc: "search"},
			{Key: "↑↓", Desc: "navigate"},
			{Key: "enter", Desc: "execute"},
			{Key: "esc", Desc: "close"},
		}
	case ModeAISummary:
		hints = []widgets.KeyHint{
			{Key: "esc", Desc: "close summary"},
		}
	case ModeHelp:
		hints = []widgets.KeyHint{
			{Key: "?/esc", Desc: "close"},
			{Key: "j/k", Desc: "scroll"},
		}
	case ModeSettings:
		hints = []widgets.KeyHint{
			{Key: "↑↓/jk", Desc: "navigate"},
			{Key: "enter", Desc: "edit"},
			{Key: "space", Desc: "toggle"},
			{Key: "ctrl+s", Desc: "save"},
			{Key: "esc", Desc: "cancel"},
		}
	case ModePR:
		hints = []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "enter", Desc: "detail"},
			{Key: "m", Desc: "merge"},
			{Key: "x", Desc: "close"},
			{Key: "r", Desc: "refresh"},
			{Key: "esc", Desc: "back"},
		}
	default:
		switch m.focused {
		case PanelFiles:
			hints = []widgets.KeyHint{
				{Key: "↑↓", Desc: "navigate"},
				{Key: "space", Desc: "stage"},
				{Key: "a", Desc: "stage all"},
				{Key: "c", Desc: "commit"},
				{Key: "P", Desc: "PRs"},
				{Key: "z", Desc: "stash"},
				{Key: "S", Desc: "staged diff"},
			}
		case PanelDiff:
			hints = []widgets.KeyHint{
				{Key: "[/]", Desc: "hunk"},
				{Key: "space", Desc: "stage hunk"},
				{Key: "u", Desc: "unstage hunk"},
				{Key: "A", Desc: "AI summary"},
				{Key: "↑↓", Desc: "scroll"},
			}
		case PanelCommits:
			hints = []widgets.KeyHint{
				{Key: "↑↓", Desc: "navigate"},
				{Key: "g", Desc: "graph"},
				{Key: "b", Desc: "blame"},
				{Key: "R", Desc: "reflog"},
				{Key: "P", Desc: "PRs"},
			}
		case PanelBranches:
			hints = []widgets.KeyHint{
				{Key: "↑↓", Desc: "navigate"},
				{Key: "enter", Desc: "checkout"},
				{Key: "n", Desc: "new"},
				{Key: "D", Desc: "delete"},
				{Key: "P/p", Desc: "push/pull"},
			}
		default:
			hints = []widgets.KeyHint{
				{Key: "tab", Desc: "panel"},
				{Key: "r", Desc: "refresh"},
			}
		}
		hints = append(hints,
			widgets.KeyHint{Key: "/", Desc: "search"},
			widgets.KeyHint{Key: "ctrl+p", Desc: "palette"},
			widgets.KeyHint{Key: "O", Desc: "settings"},
			widgets.KeyHint{Key: "?", Desc: "help"},
			widgets.KeyHint{Key: "q", Desc: "quit"},
		)
	}

	sb := widgets.NewStatusBar(m.width)
	sb.Hints = hints
	sb.Extra = m.statusLine()
	return sb.View()
}

func (m Model) statusLine() string {
	if m.statusErr {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed)).Bold(true).Render("  ✗ " + m.statusMsg)
	}
	if m.statusMsg != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Render("  ✓ " + m.statusMsg)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render("  " + m.files.Stats())
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func panelColors(focused bool) (border, title lipgloss.Color) {
	if focused {
		return lipgloss.Color("#89dceb"), lipgloss.Color("#89dceb") // sky blue — brighter when active
	}
	return lipgloss.Color("#313244"), lipgloss.Color("#6c7086") // subtle when inactive
}

func (m *Model) cyclePanel(dir int) {
	main := []PanelID{PanelFiles, PanelBranches, PanelCommits, PanelDiff}
	cur := 0
	for i, p := range main {
		if p == m.focused {
			cur = i
			break
		}
	}
	m.focused = main[(cur+dir+len(main))%len(main)]
}

func (m *Model) setStatus(msg string, isErr bool) {
	m.statusMsg = msg
	m.statusErr = isErr
}

func (m *Model) rebuildPanels() {
	layout := ComputeLayout(m.width, m.height)
	panelH, leftW, rightW := layout.LeftHeight, layout.LeftWidth, layout.RightWidth
	tabH := panelH / 3
	if tabH < 5 {
		tabH = 5
	}

	m.files = panels.NewFilesModel(leftW, tabH)
	m.branches = panels.NewBranchModel(leftW, tabH)
	m.commits = panels.NewCommitModel(leftW, panelH-2*tabH)
	m.diff = panels.NewDiffModel(rightW, panelH)
	m.commitMsg = panels.NewCommitMsgModel(leftW, panelH)
	m.stash = panels.NewStashModel(m.width-2, panelH)
	m.blame = panels.NewBlameModel(m.width-2, panelH)
	m.reflog = panels.NewReflogModel(m.width-2, panelH)
	m.pr = panels.NewPRModel(m.width-2, panelH)
	m.palette = panels.NewPaletteModel(m.width, m.height)
	m.palette.SetEntries(m.buildPaletteEntries())
	m.help = widgets.NewHelpOverlay(m.width, m.height)
	m.help.Sections = m.buildHelpSections()
	if m.mode == ModeSettings {
		m.settings.Resize(m.width, m.height)
	} else {
		m.settings = panels.NewSettingsModel(m.cfg, m.width, m.height)
	}
	m.searchInput.Width = m.width / 3
}

func (m Model) buildHelpSections() []widgets.HelpSection {
	km := m.keymap
	return []widgets.HelpSection{
		{Title: "Navigation", Bindings: []key.Binding{
			km.Up, km.Down, km.PageUp, km.PageDown,
			km.FocusFiles, km.FocusBranches, km.FocusCommits, km.FocusDiff,
			km.NextPanel, km.PrevPanel,
		}},
		{Title: "Files  (P: PRs, z: stash)", Bindings: []key.Binding{
			km.Stage, km.Unstage, km.StageAll, km.Discard, km.Commit,
		}},
		{Title: "Diff  ([/] hunks, A: AI summary)", Bindings: []key.Binding{
			km.ToggleDiffStaged,
		}},
		{Title: "Remote", Bindings: []key.Binding{km.Push, km.Pull, km.Fetch}},
		{Title: "Branches", Bindings: []key.Binding{km.Checkout, km.NewBranch, km.Delete}},
		{Title: "Commits  (g: graph, b: blame, R: reflog)", Bindings: []key.Binding{}},
		{Title: "Commit form  (tab: fields, ctrl+g: AI, ctrl+s: commit)", Bindings: []key.Binding{km.Escape}},
		{Title: "App  (ctrl+p: palette, O: settings)", Bindings: []key.Binding{km.Settings, km.Search, km.Refresh, km.Help, km.Quit}},
	}
}

// openPRPanel is the shared helper to switch to the PR overlay.
func (m Model) openPRPanel() (tea.Model, tea.Cmd) {
	m.prevFocused = m.focused
	m.focused = PanelPR
	m.mode = ModePR
	return m, m.loadPRs()
}

// handlePaletteKey handles input while the command palette is open.
func (m Model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil
	case "enter":
		e := m.palette.Current()
		if e != nil {
			m.mode = ModeNormal
			return m.executePaletteAction(e.ID)
		}
		m.mode = ModeNormal
		return m, nil
	case "up":
		m.palette.MoveUp()
		return m, nil
	case "down":
		m.palette.MoveDown()
		return m, nil
	}
	cmd := m.palette.Update(msg)
	return m, cmd
}

// executePaletteAction dispatches a palette action ID to the right command.
func (m Model) executePaletteAction(id string) (tea.Model, tea.Cmd) {
	switch id {
	case "stage":
		return m, m.toggleStage()
	case "stage_all":
		return m, m.stageAll()
	case "commit":
		m.mode = ModeCommit
		m.commitMsg.Focus()
		return m, nil
	case "push":
		return m, m.push()
	case "pull":
		return m, m.pull()
	case "fetch":
		return m, m.fetch()
	case "new_branch":
		m.mode = ModeNewBranch
		m.focused = PanelBranches
		m.branches.ShowNewBranchModal()
		return m, nil
	case "open_prs":
		return m.openPRPanel()
	case "open_stash":
		m.prevFocused = m.focused
		m.focused = PanelStash
		m.mode = ModeStash
		return m, m.loadStashes()
	case "open_blame":
		f := m.files.CurrentFile()
		if f != nil {
			m.prevFocused = m.focused
			m.focused = PanelBlame
			m.mode = ModeBlame
			path := f.Path
			return m, func() tea.Msg {
				lines, err := m.repo.Blame(path)
				return GitBlameMsg{Lines: lines, Err: err}
			}
		}
	case "open_reflog":
		m.prevFocused = m.focused
		m.focused = PanelReflog
		m.mode = ModeReflog
		return m, m.loadReflog()
	case "toggle_graph":
		m.focused = PanelCommits
		m.commits.ToggleGraph()
		if m.commits.ShowGraph {
			return m, m.loadGraph()
		}
		return m, nil
	case "toggle_staged":
		m.focused = PanelDiff
		m.diffStaged = !m.diffStaged
		return m, m.loadDiffForCursor()
	case "ai_commit":
		m.focused = PanelFiles
		m.mode = ModeCommit
		m.commitMsg.Focus()
		ch := make(chan tea.Msg, 64)
		m.aiChan = ch
		m.aiGenerating = true
		m.mode = ModeAIGenerating
		m.aiBuffer = ""
		m.commitMsg.SetValue("")
		aiGen, repo := m.aiGen, m.repo
		return m, generateCommitMsgCmd(aiGen, repo, ch)
	case "ai_diff_summary":
		m.mode = ModeAISummary
		m.aiSummary = "⟳ Generating AI summary…"
		return m, m.summarizeDiff()
	case "help":
		m.mode = ModeHelp
		return m, nil
	case "settings":
		m.settings.LoadConfig(m.cfg)
		m.mode = ModeSettings
		return m, nil
	case "refresh":
		return m, tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits())
	case "quit":
		return m, tea.Quit
	}
	return m, nil
}

// summarizeDiff generates an AI summary of the current diff.
func (m Model) summarizeDiff() tea.Cmd {
	if m.aiGen == nil || !m.aiGen.Available() {
		return func() tea.Msg {
			return AISummaryMsg{Text: "AI not configured — open settings (O) and add your Groq API key"}
		}
	}
	aiGen := m.aiGen
	hunks := m.diff.Hunks()
	diffStaged := m.diffStaged

	return func() tea.Msg {
		// Build diff text from loaded hunks; fall back to fresh load if empty.
		var sb strings.Builder
		if len(hunks) == 0 {
			return AISummaryMsg{Text: "No diff to summarise — select a changed file first"}
		}
		for _, h := range hunks {
			sb.WriteString(h.Header + "\n")
			for _, dl := range h.Lines {
				switch dl.Type {
				case git.DiffAdded:
					sb.WriteString("+" + dl.Content + "\n")
				case git.DiffRemoved:
					sb.WriteString("-" + dl.Content + "\n")
				case git.DiffContext:
					sb.WriteString(" " + dl.Content + "\n")
				}
			}
		}
		_ = diffStaged

		var result strings.Builder
		err := aiGen.Generate(context.Background(), ai.DiffSummaryPrompt(sb.String()), func(s string) {
			result.WriteString(s)
		})
		if err != nil {
			return AISummaryMsg{Text: "AI error: " + err.Error()}
		}
		return AISummaryMsg{Text: result.String()}
	}
}

// buildPaletteEntries returns the full list of commands for the palette.
func (m Model) buildPaletteEntries() []panels.PaletteEntry {
	return []panels.PaletteEntry{
		{ID: "stage", Label: "Stage / unstage current file", Keys: "space", Category: "Files"},
		{ID: "stage_all", Label: "Stage all changes", Keys: "a", Category: "Files"},
		{ID: "commit", Label: "Open commit form", Keys: "c", Category: "Files"},
		{ID: "push", Label: "Push to remote", Keys: "P", Category: "Remote"},
		{ID: "pull", Label: "Pull from remote", Keys: "p", Category: "Remote"},
		{ID: "fetch", Label: "Fetch all remotes", Keys: "f", Category: "Remote"},
		{ID: "new_branch", Label: "Create new branch", Keys: "n", Category: "Branches"},
		{ID: "open_prs", Label: "Open pull requests panel", Keys: "P (commits)", Category: "Forge"},
		{ID: "open_stash", Label: "Open stash panel", Keys: "z", Category: "Git"},
		{ID: "open_blame", Label: "Open blame for current file", Keys: "b", Category: "Git"},
		{ID: "open_reflog", Label: "Open reflog", Keys: "R", Category: "Git"},
		{ID: "toggle_graph", Label: "Toggle commit graph", Keys: "g", Category: "Commits"},
		{ID: "toggle_staged", Label: "Toggle staged diff view", Keys: "S", Category: "Diff"},
		{ID: "ai_commit", Label: "AI: generate commit message", Keys: "ctrl+g", Category: "AI"},
		{ID: "ai_diff_summary", Label: "AI: summarise current diff", Keys: "A", Category: "AI"},
		{ID: "settings", Label: "Open settings (API keys, tokens…)", Keys: "O", Category: "App"},
		{ID: "refresh", Label: "Refresh all panels", Keys: "r", Category: "App"},
		{ID: "help", Label: "Show help overlay", Keys: "?", Category: "App"},
		{ID: "quit", Label: "Quit sugi", Keys: "q", Category: "App"},
	}
}

// cleanCommitMsg deduplicates and cleans an AI-generated commit message.
// LLMs sometimes repeat the subject line in the body or add preamble.
func cleanCommitMsg(raw string) string {
	raw = strings.TrimSpace(raw)

	// Strip common AI preambles like "Here is..." or "Commit message:"
	for _, prefix := range []string{
		"Here is", "Here's", "Commit message:", "commit message:",
		"```", "---",
	} {
		if strings.HasPrefix(raw, prefix) {
			if idx := strings.Index(raw, "\n"); idx >= 0 {
				raw = strings.TrimSpace(raw[idx+1:])
			}
		}
	}

	// Split on \n\n into paragraphs
	paragraphs := strings.Split(raw, "\n\n")
	if len(paragraphs) == 0 {
		return raw
	}

	subject := strings.TrimSpace(paragraphs[0])
	// Enforce 72-char subject limit
	if len(subject) > 72 {
		subject = subject[:72]
	}

	if len(paragraphs) == 1 {
		return subject
	}

	// Collect unique body paragraphs — skip only exact duplicates of the subject
	seen := map[string]bool{subject: true}
	var bodyParts []string
	for _, p := range paragraphs[1:] {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		// Only skip if paragraph IS the subject verbatim (exact match)
		if p == subject {
			continue
		}
		seen[p] = true
		bodyParts = append(bodyParts, p)
	}

	if len(bodyParts) == 0 {
		return subject
	}
	// Use only the first body paragraph to keep messages concise
	return subject + "\n\n" + bodyParts[0]
}
