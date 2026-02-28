package ui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

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
type GitTagsMsg struct {
	Tags []git.Tag
	Err  error
}
type GitWorktreesMsg struct {
	Worktrees []git.Worktree
	Err       error
}
type GitFileHistoryMsg struct {
	Path    string
	Commits []git.Commit
	Err     error
}
type GitRebaseTodoMsg struct {
	Entries  []git.RebaseTodoEntry
	TodoPath string
	Err      error
}
type GitConflictMsg struct {
	Path   string
	Blocks []git.ConflictBlock
	Err    error
}
type GitRemotesMsg struct {
	Remotes []git.RemoteEntry
	Err     error
}
type GitBisectMsg struct {
	Status git.BisectStatus
	Err    error
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

// TickMsg is emitted by the auto-refresh ticker.
type TickMsg struct{}

// ResizeMsg is emitted 50 ms after a WindowSizeMsg to debounce rapid resize events.
type ResizeMsg struct{ W, H int }

// AccountDeleteMsg requests deletion of a named account entry.
type AccountDeleteMsg struct {
	Name      string
	ForgeType string // "github" or "gitlab"
}

// GitHeadCommitMsg carries the HEAD commit subject+body for amend mode.
type GitHeadCommitMsg struct {
	Subject string
	Body    string
	Err     error
}

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
	ModeConfirm      // destructive action confirmation dialog
	ModeNewTag       // new tag name input
	ModeRenameBranch // rename branch input
	ModeReset        // waiting for s/m/h after X
	ModeWorktree     // worktrees panel
	ModeInterRebase  // interactive rebase panel
	ModeConflict     // merge conflict resolution panel
	ModeFileHistory  // file history panel
	ModeRemotes      // remotes management panel
	ModeAddRemote    // adding a new remote (two-step modal)
	ModeBisect       // git bisect panel
	ModeAccounts     // accounts management panel
	ModeAddAccount   // add-account modal within accounts panel
)

// Extended PanelIDs (beyond the core 0-4 in layout.go)
const (
	PanelStash       PanelID = 10
	PanelBlame       PanelID = 11
	PanelReflog      PanelID = 12
	PanelPR          PanelID = 13
	PanelTags        PanelID = 14
	PanelWorktree    PanelID = 15
	PanelInterRebase PanelID = 16
	PanelConflict    PanelID = 17
	PanelFileHistory PanelID = 18
	PanelRemotes     PanelID = 19
	PanelBisect      PanelID = 20
	PanelAccounts    PanelID = 21
)

// --- Model ---

type Model struct {
	repo   *git.Client
	cfg    config.Config
	forge  forge.ForgeClient
	aiGen  *ai.Generator
	keymap KeyMap

	files       panels.FilesModel
	branches    panels.BranchModel
	commits     panels.CommitModel
	diff        panels.DiffModel
	commitMsg   panels.CommitMsgModel
	stash       panels.StashModel
	blame       panels.BlameModel
	reflog      panels.ReflogModel
	pr          panels.PRModel
	tags        panels.TagsModel
	worktree    panels.WorktreeModel
	rebase      panels.RebaseModel
	conflict    panels.ConflictModel
	fileHistory panels.CommitModel
	remotes     panels.RemotesModel
	bisect      panels.BisectModel
	accounts    panels.AccountsModel
	palette     panels.PaletteModel
	settings    panels.SettingsModel

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

	statusMsg  string
	statusErr  bool
	loading    bool
	loadingMsg string

	aiGenerating bool
	aiBuffer     string
	aiChan       chan tea.Msg

	aiSummary string // AI diff summary text

	// Confirm modal for destructive actions
	confirmModal  widgets.Modal
	confirmAction func() tea.Cmd

	// Reset mode: tracks ref target while waiting for s/m/h key
	resetTargetRef string

	// File history: path of file being browsed
	fileHistoryPath string

	// Toast notifications
	toast widgets.ToastQueue

	// Amend mode: next commit uses git commit --amend
	amendMode bool

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
	return tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits(), m.tickCmd())
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(_ time.Time) tea.Msg { return TickMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
			return ResizeMsg{W: msg.Width, H: msg.Height}
		})
	case ResizeMsg:
		if msg.W != m.width || msg.H != m.height {
			return m, nil // superseded by a later resize
		}
		m.rebuildPanels()
		return m, nil
	case TickMsg:
		return m, tea.Batch(m.loadStatus(), m.loadBranches(), m.tickCmd())
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	case panels.PaletteSelectMsg:
		m.mode = ModeNormal
		return m.executePaletteAction(msg.ID)
	case SettingsSavedMsg:
		m.aiGen = ai.NewGenerator(msg.Cfg.GroqAPIKey, msg.Cfg.GroqModel)
		m.setStatus("Settings saved ✓", false)
	default:
		return m.handleDataMsg(msg)
	}
	return m, nil
}

// handleDataMsg handles all async data messages (git results, AI, status).
// Extracted from Update to keep cyclomatic complexity manageable.
func (m Model) handleDataMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case GitHeadCommitMsg:
		if msg.Err != nil {
			m.setStatus("Cannot amend: "+msg.Err.Error(), true)
		} else {
			m.amendMode = true
			m.mode = ModeCommit
			m.commitMsg.Focus()
			val := msg.Subject
			if msg.Body != "" {
				val += "\n\n" + msg.Body
			}
			m.commitMsg.SetValue(val)
			m.setStatus("Amend mode — edit message and ctrl+s to amend HEAD", false)
		}
	case GitStatusMsg:
		m.loading = false
		if msg.Err != nil {
			m.setStatus("Error: "+msg.Err.Error(), true)
		} else {
			m.files.SetFiles(msg.Files)
			// Auto-load diff for the first selected file on startup/refresh.
			if m.focused == PanelFiles || m.focused == PanelDiff {
				return m, m.loadDiffForCursor()
			}
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
	case GitTagsMsg:
		if msg.Err == nil {
			m.tags.SetTags(msg.Tags)
		} else {
			m.setStatus("Tags: "+msg.Err.Error(), true)
		}
	case GitWorktreesMsg:
		if msg.Err == nil {
			m.worktree.SetWorktrees(msg.Worktrees)
		} else {
			m.setStatus("Worktrees: "+msg.Err.Error(), true)
		}
	case GitFileHistoryMsg:
		if msg.Err == nil {
			m.fileHistoryPath = msg.Path
			m.fileHistory.SetCommits(msg.Commits)
			m.focused = PanelFileHistory
			m.mode = ModeFileHistory
		} else {
			m.setStatus("File history: "+msg.Err.Error(), true)
		}
	case GitRebaseTodoMsg:
		if msg.Err == nil {
			m.rebase.SetEntries(msg.Entries, msg.TodoPath)
			m.focused = PanelInterRebase
			m.mode = ModeInterRebase
		} else {
			m.setStatus("Rebase: "+msg.Err.Error(), true)
		}
	case GitConflictMsg:
		if msg.Err == nil {
			m.conflict.SetConflicts(msg.Path, msg.Blocks)
			m.focused = PanelConflict
			m.mode = ModeConflict
		} else {
			m.setStatus("Conflict: "+msg.Err.Error(), true)
		}
	case GitRemotesMsg:
		if msg.Err == nil {
			m.remotes.SetRemotes(msg.Remotes)
		} else {
			m.setStatus("Remotes: "+msg.Err.Error(), true)
		}
	case GitBisectMsg:
		if msg.Err != nil {
			m.setStatus("Bisect: "+msg.Err.Error(), true)
		} else {
			m.bisect.SetStatus(msg.Status)
		}
	case GitOperationMsg:
		return m.handleGitOperationMsg(msg)
	case AIChunkMsg, AIDoneMsg, AIErrorMsg, AISummaryMsg:
		return m.handleAIMsg(msg)
	case StatusMsg:
		if msg.Text != "" {
			m.setStatus(msg.Text, msg.IsErr)
		}
	case AccountDeleteMsg:
		m.accounts.DeleteByName(&m.cfg, msg.Name, msg.ForgeType)
		if m.focused == PanelAccounts {
			m.mode = ModeAccounts
		}
		cfg := m.cfg
		return m, func() tea.Msg {
			if err := config.Save(cfg); err != nil {
				return StatusMsg{Text: "Account deleted (config write failed: " + err.Error() + ")", IsErr: true}
			}
			return StatusMsg{Text: "Account deleted ✓", IsErr: false}
		}
	}
	return m, nil
}

func (m Model) handleAIMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	case AISummaryMsg:
		if msg.Text == "" {
			m.aiSummary = "AI returned empty summary — check your AI backend"
		} else {
			m.aiSummary = msg.Text
		}
		m.mode = ModeAISummary
	}
	return m, nil
}

func (m Model) handleGitOperationMsg(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.amendMode = false
	if msg.Err != nil {
		m.setStatus(msg.Op+" failed: "+msg.Err.Error(), true)
	} else if msg.Op == "Commit" || msg.Op == "Amend" {
		m.setStatus("✓ "+msg.Op+" done  —  push with shift+P", false)
	} else {
		m.setStatus(msg.Op+" ✓", false)
	}
	cmds := []tea.Cmd{m.loadStatus(), m.loadBranches(), m.loadCommits()}
	if m.focused == PanelRemotes {
		cmds = append(cmds, m.loadRemotes())
	}
	if m.focused == PanelBisect || m.focused == PanelWorktree {
		cmds = append(cmds, m.loadBisectStatus())
	}
	return m, tea.Batch(cmds...)
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
	case ModeConfirm, ModeReset:
		return m.renderCentered(m.confirmModal.View())
	}
	return base
}

// ─── Key handling ────────────────────────────────────────────────────────────

// handleModeOverlay handles keys when a full-screen overlay mode is active.
// Returns (model, cmd, true) if the message was consumed; (m, nil, false) otherwise.
func (m Model) handleModeOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch m.mode {
	case ModeHelp:
		switch msg.String() {
		case "?", "esc":
			m.mode = ModeNormal
		case "j", "down":
			m.help.ScrollDown()
		case "k", "up":
			m.help.ScrollUp()
		}
		return m, nil, true
	case ModePalette:
		out, cmd := m.handlePaletteKey(msg)
		return out, cmd, true
	case ModeAISummary:
		if key.Matches(msg, m.keymap.Escape) {
			m.mode = ModeNormal
			m.aiSummary = ""
		}
		return m, nil, true
	case ModeConfirm:
		out, cmd := m.handleConfirmKey(msg)
		return out, cmd, true
	case ModeSettings:
		out, cmd := m.handleSettingsKey(msg)
		return out, cmd, true
	case ModeSearch:
		out, cmd := m.handleSearchKey(msg)
		return out, cmd, true
	case ModeCommit, ModeAIGenerating:
		out, cmd := m.handleCommitKey(msg)
		return out, cmd, true
	case ModeNewBranch:
		out, cmd := m.handleNewBranchKey(msg)
		return out, cmd, true
	case ModeRenameBranch:
		out, cmd := m.handleRenameBranchKey(msg)
		return out, cmd, true
	case ModeReset:
		out, cmd := m.handleResetKey(msg)
		return out, cmd, true
	case ModeNewTag:
		out, cmd := m.handleTagsKey(msg)
		return out, cmd, true
	case ModeAddRemote:
		out, cmd := m.handleRemotesKey(msg)
		return out, cmd, true
	case ModeAddAccount:
		out, cmd := m.handleAccountsKey(msg)
		return out, cmd, true
	}
	return m, nil, false
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Mode overlays intercept input before panel routing.
	if out, cmd, handled := m.handleModeOverlay(msg); handled {
		return out, cmd
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
	case key.Matches(msg, m.keymap.Refresh) && m.focused != PanelBranches:
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
	case key.Matches(msg, m.keymap.Tags):
		// Open tags panel from anywhere (except when already in it)
		if m.focused != PanelTags {
			m.prevFocused = m.focused
			m.focused = PanelTags
			return m, m.loadTags()
		}
	case key.Matches(msg, m.keymap.Worktree):
		if m.focused != PanelWorktree {
			m.prevFocused = m.focused
			m.focused = PanelWorktree
			m.mode = ModeWorktree
			return m, m.loadWorktrees()
		}
	case key.Matches(msg, m.keymap.Remotes):
		if m.focused != PanelRemotes {
			m.prevFocused = m.focused
			m.focused = PanelRemotes
			m.mode = ModeRemotes
			return m, m.loadRemotes()
		}
	case key.Matches(msg, m.keymap.Bisect):
		if m.focused != PanelBisect {
			m.prevFocused = m.focused
			m.focused = PanelBisect
			m.mode = ModeBisect
			return m, m.loadBisectStatus()
		}
	// A opens accounts panel from any panel.
	case msg.String() == "A":
		if m.focused != PanelAccounts {
			m.prevFocused = m.focused
			m.focused = PanelAccounts
			m.mode = ModeAccounts
			m.accounts.LoadConfig(m.cfg)
		}
		return m, nil
	}

	return m.dispatchPanelKey(msg)
}

func (m Model) dispatchPanelKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.focused {
	case PanelStash:
		return m.handleStashKey(msg)
	case PanelBlame:
		return m.handleBlameKey(msg)
	case PanelReflog:
		return m.handleReflogKey(msg)
	case PanelPR:
		return m.handlePRKey(msg)
	case PanelTags:
		return m.handleTagsKey(msg)
	case PanelWorktree:
		return m.handleWorktreeKey(msg)
	case PanelInterRebase:
		return m.handleInterRebaseKey(msg)
	case PanelConflict:
		return m.handleConflictKey(msg)
	case PanelFileHistory:
		return m.handleFileHistoryKey(msg)
	case PanelRemotes:
		return m.handleRemotesKey(msg)
	case PanelBisect:
		return m.handleBisectKey(msg)
	case PanelAccounts:
		return m.handleAccountsKey(msg)
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
	case PanelStash, PanelBlame, PanelReflog, PanelPR, PanelTags,
		PanelWorktree, PanelInterRebase, PanelConflict, PanelFileHistory,
		PanelRemotes, PanelBisect, PanelAccounts:
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
	case msg.String() == "ctrl+ ":
		// Multi-select toggle
		m.files.ToggleSelect()
		m.files.MoveDown()
		return m, nil
	case key.Matches(msg, m.keymap.Stage):
		if m.files.HasSelection() {
			return m, m.stageSelected()
		}
		return m, m.toggleStage()
	case key.Matches(msg, m.keymap.StageAll):
		return m, m.stageAll()
	case key.Matches(msg, m.keymap.Discard):
		f := m.files.CurrentFile()
		if f == nil {
			return m, nil
		}
		return m.confirmDiscardFile(f.Path)
	case key.Matches(msg, m.keymap.Commit):
		m.mode = ModeCommit
		m.commitMsg.Focus()
		return m, nil
	case msg.String() == "ctrl+a":
		// Amend HEAD commit — load its message and open commit form
		return m, m.loadHeadCommit()
	case key.Matches(msg, m.keymap.Push):
		m.loading = true
		m.loadingMsg = "Pushing…"
		m.setStatus("Pushing to remote…", false)
		return m, m.push()
	case key.Matches(msg, m.keymap.Pull):
		m.loading = true
		m.loadingMsg = "Pulling…"
		m.setStatus("Pulling from remote…", false)
		return m, m.pull()
	case key.Matches(msg, m.keymap.Fetch):
		m.loading = true
		m.loadingMsg = "Fetching…"
		m.setStatus("Fetching remotes…", false)
		return m, m.fetch()
	case msg.String() == "F":
		// Force push with lease (safe force push for rebase workflows)
		m.loading = true
		m.setStatus("Force pushing (with-lease)…", false)
		return m, m.forcePushWithLease()
	case key.Matches(msg, m.keymap.ToggleDiffStaged):
		m.diffStaged = !m.diffStaged
		return m, m.loadDiffForCursor()
	case key.Matches(msg, m.keymap.FileHistory):
		f := m.files.CurrentFile()
		if f != nil {
			path := f.Path
			m.prevFocused = m.focused
			return m, func() tea.Msg {
				commits, err := m.repo.LogFile(path, 100)
				return GitFileHistoryMsg{Path: path, Commits: commits, Err: err}
			}
		}
	case msg.String() == "z":
		m.prevFocused = m.focused
		m.focused = PanelStash
		m.mode = ModeStash
		return m, m.loadStashes()
	case msg.String() == "Z":
		// Stash all current changes
		m.setStatus("Stashing changes…", false)
		return m, func() tea.Msg {
			return GitOperationMsg{Op: "Stash", Err: m.repo.StashPush("")}
		}
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
		if b == nil || b.IsCurrent {
			return m, nil
		}
		return m.confirmBranchDelete(b.Name)
	case key.Matches(msg, m.keymap.Push):
		m.loading = true
		m.loadingMsg = "Pushing…"
		m.setStatus("Pushing to remote…", false)
		return m, m.push()
	case key.Matches(msg, m.keymap.Pull):
		m.loading = true
		m.loadingMsg = "Pulling…"
		m.setStatus("Pulling from remote…", false)
		return m, m.pull()
	case msg.String() == "P":
		return m.openPRPanel()
	case key.Matches(msg, m.keymap.Merge):
		b := m.branches.CurrentBranch()
		if b == nil || b.IsCurrent {
			return m, nil
		}
		if m.repo.MergeInProgress() {
			return m, func() tea.Msg {
				return GitOperationMsg{Op: "Merge abort", Err: m.repo.MergeAbort()}
			}
		}
		return m.confirmBranchMerge(b.Name)
	case key.Matches(msg, m.keymap.Rebase):
		b := m.branches.CurrentBranch()
		if b == nil || b.IsCurrent {
			return m, nil
		}
		if m.repo.RebaseInProgress() {
			return m, func() tea.Msg {
				return GitOperationMsg{Op: "Rebase abort", Err: m.repo.RebaseAbort()}
			}
		}
		return m.confirmBranchRebase(b.Name)
	case key.Matches(msg, m.keymap.RenameBranch):
		b := m.branches.CurrentBranch()
		if b == nil {
			return m, nil
		}
		m.mode = ModeRenameBranch
		m.branches.ShowRenameBranchModal(b.Name)
		return m, nil
	case key.Matches(msg, m.keymap.OpenBrowser):
		b := m.branches.CurrentBranch()
		if b != nil {
			return m, m.openBranchInBrowser(b.Name)
		}
		return m, nil
	case key.Matches(msg, m.keymap.Refresh):
		m.loading = true
		m.loadingMsg = "Refreshing…"
		return m, tea.Batch(m.loadStatus(), m.loadBranches(), m.loadCommits())
	}
	return m, nil
}

func (m Model) confirmBranchDelete(name string) (tea.Model, tea.Cmd) {
	m.confirmModal = widgets.NewConfirmModal(
		"Delete branch?",
		"Delete local branch '"+name+"'?\n\nRemote branch is not affected.",
	)
	m.confirmModal.Show()
	m.mode = ModeConfirm
	m.confirmAction = func() tea.Cmd { return m.deleteBranch(name) }
	return m, nil
}

func (m Model) confirmBranchMerge(name string) (tea.Model, tea.Cmd) {
	m.confirmModal = widgets.NewConfirmModal(
		"Merge branch?",
		"Merge '"+name+"' into current branch.",
	)
	m.confirmModal.Show()
	m.mode = ModeConfirm
	m.confirmAction = func() tea.Cmd {
		return func() tea.Msg {
			return GitOperationMsg{Op: "Merge " + name, Err: m.repo.MergeBranch(name)}
		}
	}
	return m, nil
}

func (m Model) confirmBranchRebase(name string) (tea.Model, tea.Cmd) {
	m.confirmModal = widgets.NewConfirmModal(
		"Rebase onto branch?",
		"Rebase current branch onto '"+name+"'.",
	)
	m.confirmModal.Show()
	m.mode = ModeConfirm
	m.confirmAction = func() tea.Cmd {
		return func() tea.Msg {
			return GitOperationMsg{Op: "Rebase onto " + name, Err: m.repo.RebaseBranch(name)}
		}
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
	case msg.String() == "y":
		// Copy commit hash to clipboard
		c := m.commits.CurrentCommit()
		if c != nil {
			return m, m.copyToClipboard(c.Hash, c.ShortHash)
		}
	case msg.String() == "ctrl+a":
		// Amend HEAD commit — only makes sense for the top commit
		return m, m.loadHeadCommit()
	case key.Matches(msg, m.keymap.CherryPick):
		c := m.commits.CurrentCommit()
		if c != nil {
			return m.confirmCherryPick(c.Hash, c.ShortHash, c.Subject)
		}
		return m, nil
	case key.Matches(msg, m.keymap.Revert):
		c := m.commits.CurrentCommit()
		if c != nil {
			return m.confirmRevert(c.Hash, c.ShortHash, c.Subject)
		}
		return m, nil
	case key.Matches(msg, m.keymap.Reset):
		c := m.commits.CurrentCommit()
		if c != nil {
			m.resetTargetRef = c.Hash
			m.confirmModal = widgets.NewConfirmModal(
				"Reset HEAD to "+c.ShortHash+"?",
				"Choose reset type:\n  [s] soft  — keep staged\n  [m] mixed — unstage changes\n  [h] hard  — discard all",
			)
			m.confirmModal.Show()
			m.mode = ModeReset
		}
		return m, nil
	case key.Matches(msg, m.keymap.InterRebase):
		c := m.commits.CurrentCommit()
		if c != nil {
			hash := c.Hash
			return m, func() tea.Msg {
				entries, todoPath, err := m.repo.StartInteractiveRebase(hash)
				return GitRebaseTodoMsg{Entries: entries, TodoPath: todoPath, Err: err}
			}
		}
		return m, nil
	case key.Matches(msg, m.keymap.OpenBrowser):
		c := m.commits.CurrentCommit()
		if c != nil && m.forge != nil {
			return m, m.openCommitInBrowser(c.Hash)
		}
		return m, nil
	case key.Matches(msg, m.keymap.Tags):
		m.prevFocused = m.focused
		m.focused = PanelTags
		return m, m.loadTags()
	}
	return m, nil
}

func (m Model) confirmDiscardFile(path string) (tea.Model, tea.Cmd) {
	m.confirmModal = widgets.NewConfirmModal(
		"Discard changes?",
		"Permanently discard all changes to:\n  "+path+"\n\nThis cannot be undone.",
	)
	m.confirmModal.Show()
	m.mode = ModeConfirm
	m.confirmAction = func() tea.Cmd {
		return func() tea.Msg {
			return GitOperationMsg{Op: "Discard " + path, Err: m.repo.DiscardFile(path)}
		}
	}
	return m, nil
}

func (m Model) confirmCherryPick(hash, short, subject string) (tea.Model, tea.Cmd) {
	m.confirmModal = widgets.NewConfirmModal(
		"Cherry-pick commit?",
		"Apply "+short+" to current branch:\n  "+subject,
	)
	m.confirmModal.Show()
	m.mode = ModeConfirm
	m.confirmAction = func() tea.Cmd {
		return func() tea.Msg {
			return GitOperationMsg{Op: "Cherry-pick " + short, Err: m.repo.CherryPick(hash)}
		}
	}
	return m, nil
}

func (m Model) confirmRevert(hash, short, subject string) (tea.Model, tea.Cmd) {
	m.confirmModal = widgets.NewConfirmModal(
		"Revert commit?",
		"Create a new commit that undoes:\n  "+short+"  "+subject,
	)
	m.confirmModal.Show()
	m.mode = ModeConfirm
	m.confirmAction = func() tea.Cmd {
		return func() tea.Msg {
			return GitOperationMsg{Op: "Revert " + short, Err: m.repo.RevertCommit(hash)}
		}
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
	case msg.String() == "ctrl+i":
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

func (m Model) handleTagsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ModeNewTag: forward to modal input
	if m.mode == ModeNewTag {
		switch msg.String() {
		case "esc":
			m.mode = ModeNormal
			m.tags.HideModal()
		case "enter":
			name := m.tags.ModalInput()
			m.tags.HideModal()
			m.mode = ModeNormal
			if name != "" {
				return m, func() tea.Msg {
					return GitOperationMsg{Op: "Create tag " + name, Err: m.repo.CreateTag(name)}
				}
			}
		default:
			return m, m.tags.UpdateModalInput(msg)
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.tags.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.tags.MoveUp()
	case key.Matches(msg, m.keymap.NewBranch): // 'n' — new tag
		m.tags.ShowNewTagModal()
		m.mode = ModeNewTag
	case key.Matches(msg, m.keymap.Delete): // 'D' — delete tag
		t := m.tags.CurrentTag()
		if t != nil {
			name := t.Name
			m.confirmModal = widgets.NewConfirmModal(
				"Delete tag?",
				"Delete local tag '"+name+"'?\n\nThis does not affect the remote.",
			)
			m.confirmModal.Show()
			m.mode = ModeConfirm
			m.confirmAction = func() tea.Cmd {
				return func() tea.Msg {
					return GitOperationMsg{Op: "Delete tag " + name, Err: m.repo.DeleteTag(name)}
				}
			}
		}
	case key.Matches(msg, m.keymap.Push): // 'P' — push tag to origin
		t := m.tags.CurrentTag()
		if t != nil {
			name := t.Name
			m.setStatus("Pushing tag "+name+"…", false)
			return m, func() tea.Msg {
				return GitOperationMsg{Op: "Push tag " + name, Err: m.repo.PushTag("origin", name)}
			}
		}
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
		m.amendMode = false
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

func (m Model) handleRenameBranchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			return m, func() tea.Msg {
				return GitOperationMsg{Op: "Rename branch to " + name, Err: m.repo.RenameBranch(name)}
			}
		}
		return m, nil
	}
	return m, m.branches.UpdateModalInput(msg)
}

func (m Model) handleResetKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ref := m.resetTargetRef
	switch msg.String() {
	case "s":
		m.mode = ModeNormal
		m.confirmModal.Hide()
		return m, func() tea.Msg {
			return GitOperationMsg{Op: "Reset soft to " + ref[:7], Err: m.repo.ResetSoft(ref)}
		}
	case "m":
		m.mode = ModeNormal
		m.confirmModal.Hide()
		return m, func() tea.Msg {
			return GitOperationMsg{Op: "Reset mixed to " + ref[:7], Err: m.repo.ResetMixed(ref)}
		}
	case "h":
		m.mode = ModeNormal
		m.confirmModal.Hide()
		return m, func() tea.Msg {
			return GitOperationMsg{Op: "Reset hard to " + ref[:7], Err: m.repo.ResetHard(ref)}
		}
	case "esc", "q", "n":
		m.mode = ModeNormal
		m.confirmModal.Hide()
		m.setStatus("Cancelled", false)
	}
	return m, nil
}

func (m Model) handleWorktreeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.worktree.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.worktree.MoveUp()
	case key.Matches(msg, m.keymap.Delete):
		wt := m.worktree.CurrentWorktree()
		if wt != nil && !wt.IsMain {
			path := wt.Path
			m.confirmModal = widgets.NewConfirmModal(
				"Remove worktree?",
				"Remove worktree at:\n  "+path,
			)
			m.confirmModal.Show()
			m.mode = ModeConfirm
			m.confirmAction = func() tea.Cmd {
				return func() tea.Msg {
					return GitOperationMsg{Op: "Remove worktree", Err: m.repo.RemoveWorktree(path)}
				}
			}
		}
	case key.Matches(msg, m.keymap.Checkout):
		wt := m.worktree.CurrentWorktree()
		if wt != nil {
			m.setStatus("Worktree: "+wt.Path, false)
		}
	}
	return m, nil
}

func (m Model) handleInterRebaseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		_ = m.repo.RebaseAbort()
		m.mode = ModeNormal
		m.focused = m.prevFocused
		m.setStatus("Rebase aborted", false)
	case key.Matches(msg, m.keymap.Down):
		m.rebase.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.rebase.MoveUp()
	case msg.String() == "a":
		m.rebase.CycleAction()
	case msg.String() == "K":
		m.rebase.MoveEntryUp()
	case msg.String() == "J":
		m.rebase.MoveEntryDown()
	case key.Matches(msg, m.keymap.Confirm):
		todoPath := m.rebase.TodoPath
		entries := m.rebase.Entries
		return m, func() tea.Msg {
			if err := git.WriteRebaseTodo(todoPath, entries); err != nil {
				return GitOperationMsg{Op: "Interactive rebase", Err: err}
			}
			return GitOperationMsg{Op: "Interactive rebase", Err: m.repo.ContinueRebase()}
		}
	}
	return m, nil
}

func (m Model) handleConflictKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.conflict.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.conflict.MoveUp()
	case msg.String() == "o":
		path := m.conflict.FilePath
		return m, func() tea.Msg {
			if err := git.ResolveConflict(path, "ours"); err != nil {
				return GitOperationMsg{Op: "Resolve ours", Err: err}
			}
			blocks, _ := reloadConflict(path)
			return GitConflictMsg{Path: path, Blocks: blocks}
		}
	case msg.String() == "t":
		path := m.conflict.FilePath
		return m, func() tea.Msg {
			if err := git.ResolveConflict(path, "theirs"); err != nil {
				return GitOperationMsg{Op: "Resolve theirs", Err: err}
			}
			blocks, _ := reloadConflict(path)
			return GitConflictMsg{Path: path, Blocks: blocks}
		}
	case msg.String() == "O":
		path := m.conflict.FilePath
		return m, func() tea.Msg {
			if err := git.ResolveConflict(path, "ours"); err != nil {
				return GitOperationMsg{Op: "Resolve all ours", Err: err}
			}
			if err := m.repo.MarkResolved(path); err != nil {
				return GitOperationMsg{Op: "Mark resolved", Err: err}
			}
			return GitOperationMsg{Op: "Resolved with ours"}
		}
	case msg.String() == "T":
		path := m.conflict.FilePath
		return m, func() tea.Msg {
			if err := git.ResolveConflict(path, "theirs"); err != nil {
				return GitOperationMsg{Op: "Resolve all theirs", Err: err}
			}
			if err := m.repo.MarkResolved(path); err != nil {
				return GitOperationMsg{Op: "Mark resolved", Err: err}
			}
			return GitOperationMsg{Op: "Resolved with theirs"}
		}
	case key.Matches(msg, m.keymap.Confirm):
		path := m.conflict.FilePath
		return m, func() tea.Msg {
			return GitOperationMsg{Op: "Mark resolved " + path, Err: m.repo.MarkResolved(path)}
		}
	}
	return m, nil
}

func (m Model) handleFileHistoryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.fileHistory.MoveDown()
		return m, m.loadFileHistoryDiff()
	case key.Matches(msg, m.keymap.Up):
		m.fileHistory.MoveUp()
		return m, m.loadFileHistoryDiff()
	case key.Matches(msg, m.keymap.PageDown):
		m.fileHistory.PageDown()
		return m, m.loadFileHistoryDiff()
	case key.Matches(msg, m.keymap.PageUp):
		m.fileHistory.PageUp()
		return m, m.loadFileHistoryDiff()
	case key.Matches(msg, m.keymap.Confirm):
		return m, m.loadFileHistoryDiff()
	}
	return m, nil
}

func reloadConflict(path string) ([]git.ConflictBlock, error) {
	blocks, _, err := git.ConflictedFile(path)
	return blocks, err
}

func (m Model) handleRemotesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == ModeAddRemote {
		switch msg.String() {
		case "esc":
			m.mode = ModeRemotes
			m.remotes.HideModal()
		case "enter":
			val := m.remotes.ModalInput()
			if m.remotes.IsAddingURL() {
				// Confirm: add the remote
				name := m.remotes.RemoteName()
				url := val
				m.remotes.HideModal()
				m.mode = ModeRemotes
				return m, func() tea.Msg {
					return GitOperationMsg{Op: "Add remote " + name, Err: m.repo.AddRemote(name, url)}
				}
			} else if val != "" {
				// Advance to URL step
				m.remotes.AdvanceToURL(val)
			}
		default:
			return m, m.remotes.UpdateModalInput(msg)
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.remotes.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.remotes.MoveUp()
	case key.Matches(msg, m.keymap.NewBranch): // 'n' = add remote
		m.remotes.ShowAddModal()
		m.mode = ModeAddRemote
	case key.Matches(msg, m.keymap.Delete): // 'D' = remove remote
		r := m.remotes.CurrentRemote()
		if r != nil {
			name := r.Name
			m.confirmModal = widgets.NewConfirmModal(
				"Remove remote?",
				"Remove remote '"+name+"'?\n\nThis does not affect the remote server.",
			)
			m.confirmModal.Show()
			m.mode = ModeConfirm
			m.confirmAction = func() tea.Cmd {
				return func() tea.Msg {
					return GitOperationMsg{Op: "Remove remote " + name, Err: m.repo.RemoveRemote(name)}
				}
			}
		}
	case key.Matches(msg, m.keymap.Fetch): // 'f' = fetch this remote
		r := m.remotes.CurrentRemote()
		if r != nil {
			name := r.Name
			return m, func() tea.Msg {
				return GitOperationMsg{Op: "Fetch " + name, Err: m.repo.FetchRemote(name)}
			}
		}
	case key.Matches(msg, m.keymap.Checkout): // enter = show URL in status
		r := m.remotes.CurrentRemote()
		if r != nil {
			m.setStatus(r.Name+": "+r.FetchURL, false)
		}
	}
	return m, nil
}

func (m Model) handleBisectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Escape):
		m.mode = ModeNormal
		m.focused = m.prevFocused
	case key.Matches(msg, m.keymap.Down):
		m.bisect.MoveDown()
	case key.Matches(msg, m.keymap.Up):
		m.bisect.MoveUp()
	case msg.String() == "s": // start
		if !m.bisect.Status.InProgress {
			return m, func() tea.Msg {
				if err := m.repo.BisectStart(); err != nil {
					return GitOperationMsg{Op: "Bisect start", Err: err}
				}
				return GitBisectMsg{Status: m.repo.BisectGetStatus()}
			}
		}
	case msg.String() == "g": // mark good
		return m, func() tea.Msg {
			if err := m.repo.BisectGood(""); err != nil {
				return GitOperationMsg{Op: "Bisect good", Err: err}
			}
			return GitBisectMsg{Status: m.repo.BisectGetStatus()}
		}
	case msg.String() == "b": // mark bad
		return m, func() tea.Msg {
			if err := m.repo.BisectBad(""); err != nil {
				return GitOperationMsg{Op: "Bisect bad", Err: err}
			}
			return GitBisectMsg{Status: m.repo.BisectGetStatus()}
		}
	case msg.String() == "k": // skip
		return m, func() tea.Msg {
			if err := m.repo.BisectSkip(); err != nil {
				return GitOperationMsg{Op: "Bisect skip", Err: err}
			}
			return GitBisectMsg{Status: m.repo.BisectGetStatus()}
		}
	case msg.String() == "r": // reset
		return m, func() tea.Msg {
			if err := m.repo.BisectReset(); err != nil {
				return GitOperationMsg{Op: "Bisect reset", Err: err}
			}
			return GitBisectMsg{Status: m.repo.BisectGetStatus()}
		}
	}
	return m, nil
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
		branches, err := m.repo.BranchesWithTracking()
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
	case PanelFiles, PanelDiff:
		// Diff panel always shows the selected file's working-tree diff.
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

func (m Model) loadTags() tea.Cmd {
	return func() tea.Msg {
		tags, err := m.repo.Tags()
		return GitTagsMsg{Tags: tags, Err: err}
	}
}

func (m Model) loadWorktrees() tea.Cmd {
	return func() tea.Msg {
		wts, err := m.repo.Worktrees()
		return GitWorktreesMsg{Worktrees: wts, Err: err}
	}
}

func (m Model) loadRemotes() tea.Cmd {
	return func() tea.Msg {
		remotes, err := m.repo.ListRemotes()
		return GitRemotesMsg{Remotes: remotes, Err: err}
	}
}

func (m Model) loadBisectStatus() tea.Cmd {
	return func() tea.Msg {
		return GitBisectMsg{Status: m.repo.BisectGetStatus()}
	}
}

func (m Model) loadFileHistoryDiff() tea.Cmd {
	c := m.fileHistory.CurrentCommit()
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

func (m Model) openBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		if err := cmd.Run(); err != nil {
			return StatusMsg{Text: "Cannot open browser: " + err.Error(), IsErr: true}
		}
		return StatusMsg{Text: "Opened in browser"}
	}
}

func forgeWebBase(info forge.ForgeInfo) string {
	return "https://" + info.Host + "/" + info.Owner + "/" + info.Repo
}

func (m Model) openCommitInBrowser(hash string) tea.Cmd {
	if m.forge == nil {
		return func() tea.Msg {
			return StatusMsg{Text: "No forge detected — set GITHUB_TOKEN or GITLAB_TOKEN", IsErr: true}
		}
	}
	info := m.forge.ForgeInfo()
	url := forgeWebBase(info) + "/commit/" + hash
	return m.openBrowser(url)
}

func (m Model) openBranchInBrowser(branch string) tea.Cmd {
	if m.forge == nil {
		return func() tea.Msg {
			return StatusMsg{Text: "No forge detected — set GITHUB_TOKEN or GITLAB_TOKEN", IsErr: true}
		}
	}
	info := m.forge.ForgeInfo()
	url := forgeWebBase(info) + "/tree/" + branch
	return m.openBrowser(url)
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

func (m Model) stageSelected() tea.Cmd {
	files := m.files.SelectedFiles()
	if len(files) == 0 {
		return nil
	}
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	n := len(paths)
	return func() tea.Msg {
		return GitOperationMsg{Op: fmt.Sprintf("Stage %d files", n), Err: m.repo.Stage(paths...)}
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
	if m.amendMode {
		return func() tea.Msg {
			return GitOperationMsg{Op: "Amend", Err: m.repo.CommitAmend(message)}
		}
	}
	return func() tea.Msg {
		return GitOperationMsg{Op: "Commit", Err: m.repo.Commit(message)}
	}
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.mode = ModeNormal
		m.confirmModal.Hide()
		if m.confirmAction != nil {
			return m, m.confirmAction()
		}
	case "n", "esc", "q":
		m.mode = ModeNormal
		m.confirmModal.Hide()
		m.setStatus("Cancelled", false)
	}
	return m, nil
}

func (m Model) loadHeadCommit() tea.Cmd {
	return func() tea.Msg {
		c, err := m.repo.CommitDetail("HEAD")
		if err != nil || c == nil {
			return GitHeadCommitMsg{Err: fmt.Errorf("could not load HEAD commit")}
		}
		return GitHeadCommitMsg{Subject: c.Subject, Body: c.Body}
	}
}

func (m Model) forcePushWithLease() tea.Cmd {
	repo := m.repo
	return func() tea.Msg {
		branch, err := repo.CurrentBranch()
		if err != nil {
			return GitOperationMsg{Op: "Force push", Err: err}
		}
		if err := repo.PushForceWithLease("origin", branch); err != nil {
			return GitOperationMsg{Op: "Force push", Err: err}
		}
		return GitOperationMsg{Op: "Force push " + branch + " → origin"}
	}
}

func (m Model) copyToClipboard(hash, short string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("pbcopy")
		case "windows":
			cmd = exec.Command("clip")
		default:
			if _, err := exec.LookPath("xclip"); err == nil {
				cmd = exec.Command("xclip", "-selection", "clipboard")
			} else {
				cmd = exec.Command("xsel", "--clipboard", "--input")
			}
		}
		cmd.Stdin = strings.NewReader(hash)
		if err := cmd.Run(); err != nil {
			return StatusMsg{Text: "Clipboard unavailable: " + err.Error(), IsErr: true}
		}
		return StatusMsg{Text: "Copied " + short + " to clipboard"}
	}
}

func (m Model) push() tea.Cmd {
	repo := m.repo
	return func() tea.Msg {
		// Check if current branch has an upstream set.
		branch, err := repo.CurrentBranch()
		if err != nil {
			return GitOperationMsg{Op: "Push", Err: err}
		}
		// Try a normal push first.
		pushErr := repo.Push()
		if pushErr == nil {
			return GitOperationMsg{Op: "Push " + branch}
		}
		// If no upstream, auto-set it to origin/<branch>.
		errStr := pushErr.Error()
		if strings.Contains(errStr, "no upstream") ||
			strings.Contains(errStr, "set-upstream") ||
			strings.Contains(errStr, "has no upstream") {
			setErr := repo.PushSetUpstream("origin", branch)
			if setErr != nil {
				return GitOperationMsg{Op: "Push " + branch, Err: setErr}
			}
			return GitOperationMsg{Op: "Push " + branch + " → origin"}
		}
		return GitOperationMsg{Op: "Push", Err: pushErr}
	}
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
	bg := lipgloss.Color("#181825")
	repoStyle := lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("#89b4fa")).Bold(true)
	branchStyle := lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("#a6e3a1"))
	mutedStyle := lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("#45475a"))

	repoName := repoStyle.Render(" ⬡ " + m.repo.RepoName())
	branch, _ := m.repo.CurrentBranch()
	sep := mutedStyle.Render("  ⎇ ")
	branchStr := branchStyle.Render(branch)

	extra := ""
	if m.loading {
		extra += lipgloss.NewStyle().Background(bg).
			Foreground(lipgloss.Color("#f9e2af")).Render("   ⟳ " + m.loadingMsg)
	}
	if m.mode == ModeSearch {
		extra += lipgloss.NewStyle().Background(bg).
			Foreground(lipgloss.Color("#cba6f7")).Render("   / " + m.searchInput.Value())
	}
	if m.aiGenerating {
		extra += lipgloss.NewStyle().Background(bg).
			Foreground(lipgloss.Color("#94e2d5")).Render("   ✦ AI…")
	}
	// In-progress git operation badges
	if m.repo.MergeInProgress() {
		extra += "  " + lipgloss.NewStyle().
			Background(lipgloss.Color("#f9e2af")).Foreground(lipgloss.Color("#1e1e2e")).
			Bold(true).Padding(0, 1).Render("MERGE") +
			lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("#585b70")).Render(" m:abort")
	}
	if m.repo.RebaseInProgress() {
		extra += "  " + lipgloss.NewStyle().
			Background(lipgloss.Color("#cba6f7")).Foreground(lipgloss.Color("#1e1e2e")).
			Bold(true).Padding(0, 1).Render("REBASE") +
			lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("#585b70")).Render(" esc:abort")
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
	barStyle := lipgloss.NewStyle().Background(bg)
	return barStyle.Width(m.width).Render(left + strings.Repeat(" ", pad) + right)
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
	case PanelTags:
		tagHint := "TAGS  n:new  D:delete  P:push to remote  esc:close"
		if m.mode == ModeNewTag {
			return m.renderCentered(m.tags.ModalView())
		}
		return m.renderOverlay(tagHint, m.tags.View(), layout)
	case PanelWorktree:
		hint := "WORKTREES  enter:show path  D:remove  esc:close"
		return m.renderOverlay(hint, m.worktree.View(), layout)
	case PanelInterRebase:
		hint := "INTERACTIVE REBASE  a:cycle action  K/J:reorder  enter:apply  esc:abort"
		return m.renderOverlay(hint, m.rebase.View(), layout)
	case PanelConflict:
		hint := "CONFLICTS  o:ours  t:theirs  O:all ours  T:all theirs  enter:mark resolved  esc:close"
		return m.renderOverlay(hint, m.conflict.View(), layout)
	case PanelFileHistory:
		fname := m.fileHistoryPath
		hint := "HISTORY: " + fname + "  ↑↓:navigate  enter:show diff  esc:close"
		return m.renderOverlay(hint, m.fileHistory.View(), layout)
	case PanelRemotes:
		hint := "REMOTES  n:add  D:remove  f:fetch  enter:show URL  esc:close"
		if m.mode == ModeAddRemote {
			return m.renderCentered(m.remotes.ModalView())
		}
		return m.renderOverlay(hint, m.remotes.View(), layout)
	case PanelBisect:
		bisectStatus := ""
		if m.bisect.Status.InProgress {
			bisectStatus = "  g:good  b:bad  k:skip  r:reset"
		} else {
			bisectStatus = "  s:start"
		}
		hint := "GIT BISECT" + bisectStatus + "  esc:close"
		return m.renderOverlay(hint, m.bisect.View(), layout)
	case PanelAccounts:
		if m.mode == ModeAddAccount {
			return m.renderCentered(m.accounts.ModalView())
		}
		hint := "ACCOUNTS  tab:switch  enter:activate  n:add  D:delete  esc:close"
		return m.renderOverlay(hint, m.accounts.View(), layout)
	}

	if m.mode == ModeCommit || m.mode == ModeAIGenerating {
		return m.renderCommitMode(layout)
	}

	panelH, leftW, rightW := layout.LeftHeight, layout.LeftWidth, layout.RightWidth
	tabH, commitH := splitPanelHeight(panelH)

	left := lipgloss.JoinVertical(lipgloss.Left,
		m.renderFilesPanel(leftW, tabH),
		m.renderBranchesPanel(leftW, tabH),
		m.renderCommitsPanel(leftW, commitH),
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
	selStr := ""
	if m.files.HasSelection() {
		n := m.files.SelectionCount()
		selStr = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Bold(true).
			Render(fmt.Sprintf("  ◆ %d selected", n))
	}
	title := lipgloss.NewStyle().Foreground(tc).Bold(focused).Render(" ▸ FILES") + statsStr + selStr +
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOverlay)).Render("  [1]")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(bc).
		Width(width - 2).Height(height - 2).
		Render(title + "\n" + m.files.View())
}

func (m Model) renderBranchesPanel(width, height int) string {
	focused := m.focused == PanelBranches && m.mode == ModeNormal
	bc, tc := panelColors(focused)

	mergeHint := ""
	if m.repo.MergeInProgress() {
		mergeHint = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow)).Render("  ⚡ merge")
	} else if m.repo.RebaseInProgress() {
		mergeHint = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple)).Render("  ⚡ rebase")
	}

	title := lipgloss.NewStyle().Foreground(tc).Bold(focused).Render(" ⎇ BRANCHES") + mergeHint +
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOverlay)).Render("  [2]")
	inner := m.branches.View()
	if m.mode == ModeNewBranch || m.mode == ModeRenameBranch {
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
	title := lipgloss.NewStyle().Foreground(tc).Bold(focused).Render(" ● COMMITS") + extra
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(bc).
		Width(width - 2).Height(height - 2).
		Render(title + "\n" + m.commits.View())
}

func (m Model) renderDiffPanel(width, height int) string {
	focused := m.focused == PanelDiff
	bc, tc := panelColors(focused)
	extra := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOverlay)).Render("  [4]")

	// Show context: file diff vs commit diff
	if m.focused == PanelCommits {
		c := m.commits.CurrentCommit()
		if c != nil {
			extra = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).
				Render("  "+c.ShortHash) + extra
		}
	} else {
		// File diff context
		if m.diffStaged {
			extra = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Render("  staged") + extra
		} else {
			extra = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render("  unstaged") + extra
		}
		if m.currentFilePath != "" {
			fname := m.currentFilePath
			maxFnameW := width/2 - 10
			if len(fname) > maxFnameW && maxFnameW > 4 {
				fname = "…" + fname[len(fname)-maxFnameW+1:]
			}
			extra = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTeal)).Render("  "+fname) + extra
		}
	}
	if m.diff.HunkCount() > 0 {
		extra += lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).
			Render("  " + m.diff.ScrollInfo())
	}
	title := lipgloss.NewStyle().Foreground(tc).Bold(focused).Render(" ≋ DIFF") + extra
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(bc).
		Width(width - 2).Height(height - 2).
		Render(title + "\n" + m.diff.View())
}

func (m Model) focusedHints() []widgets.KeyHint {
	switch m.focused {
	case PanelFiles:
		return []widgets.KeyHint{
			{Key: "space", Desc: "stage"},
			{Key: "a", Desc: "stage all"},
			{Key: "c", Desc: "commit"},
			{Key: "L", Desc: "file history"},
			{Key: "ctrl+a", Desc: "amend"},
			{Key: "P", Desc: "push"},
			{Key: "p", Desc: "pull"},
			{Key: "Z", Desc: "stash"},
			{Key: "d", Desc: "discard"},
		}
	case PanelDiff:
		return []widgets.KeyHint{
			{Key: "[/]", Desc: "hunk"},
			{Key: "space", Desc: "stage hunk"},
			{Key: "u", Desc: "unstage hunk"},
			{Key: "ctrl+i", Desc: "AI summary"},
			{Key: "↑↓", Desc: "scroll"},
		}
	case PanelCommits:
		return []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "y", Desc: "copy hash"},
			{Key: "C", Desc: "cherry-pick"},
			{Key: "v", Desc: "revert"},
			{Key: "X", Desc: "reset"},
			{Key: "i", Desc: "interactive rebase"},
			{Key: "o", Desc: "browser"},
			{Key: "ctrl+a", Desc: "amend HEAD"},
			{Key: "g", Desc: "graph"},
			{Key: "b", Desc: "blame"},
		}
	case PanelTags:
		return []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "n", Desc: "new tag"},
			{Key: "D", Desc: "delete"},
			{Key: "P", Desc: "push"},
			{Key: "esc", Desc: "close"},
		}
	case PanelRemotes:
		return []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "n", Desc: "add"},
			{Key: "D", Desc: "remove"},
			{Key: "f", Desc: "fetch"},
			{Key: "enter", Desc: "show URL"},
			{Key: "esc", Desc: "close"},
		}
	case PanelBisect:
		if m.bisect.Status.InProgress {
			return []widgets.KeyHint{
				{Key: "g", Desc: "good"},
				{Key: "b", Desc: "bad"},
				{Key: "k", Desc: "skip"},
				{Key: "r", Desc: "reset"},
				{Key: "esc", Desc: "close"},
			}
		}
		return []widgets.KeyHint{
			{Key: "s", Desc: "start bisect"},
			{Key: "esc", Desc: "close"},
		}
	case PanelWorktree:
		return []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "enter", Desc: "show path"},
			{Key: "D", Desc: "remove"},
			{Key: "esc", Desc: "close"},
		}
	case PanelInterRebase:
		return []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "a", Desc: "cycle action"},
			{Key: "K/J", Desc: "reorder"},
			{Key: "enter", Desc: "apply"},
			{Key: "esc", Desc: "abort"},
		}
	case PanelConflict:
		return []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "o/t", Desc: "ours/theirs"},
			{Key: "O/T", Desc: "all ours/theirs"},
			{Key: "enter", Desc: "mark resolved"},
			{Key: "esc", Desc: "close"},
		}
	case PanelFileHistory:
		return []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "enter", Desc: "show diff"},
			{Key: "esc", Desc: "close"},
		}
	case PanelBranches:
		return []widgets.KeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "enter", Desc: "checkout"},
			{Key: "n", Desc: "new"},
			{Key: "R", Desc: "rename"},
			{Key: "D", Desc: "delete"},
			{Key: "m", Desc: "merge"},
			{Key: "r", Desc: "rebase"},
			{Key: "o", Desc: "browser"},
			{Key: "P/p", Desc: "push/pull"},
		}
	case PanelAccounts:
		return []widgets.KeyHint{
			{Key: "tab", Desc: "switch forge"},
			{Key: "enter", Desc: "activate"},
			{Key: "n", Desc: "add"},
			{Key: "D", Desc: "delete"},
			{Key: "esc", Desc: "close"},
		}
	default:
		return []widgets.KeyHint{
			{Key: "tab", Desc: "panel"},
			{Key: "r", Desc: "refresh"},
		}
	}
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
		hints = append(m.focusedHints(),
			widgets.KeyHint{Key: "/", Desc: "search"},
			widgets.KeyHint{Key: "ctrl+p", Desc: "palette/alt+p"},
			widgets.KeyHint{Key: "O", Desc: "settings"},
			widgets.KeyHint{Key: "?", Desc: "help"},
			widgets.KeyHint{Key: "q", Desc: "quit"},
		)
	}

	sb := widgets.NewStatusBar(m.width)
	sb.Hints = hints
	sb.Extra = m.statusLine()
	// Compose mode pills
	var pill string
	if m.repo.BisectInProgress() {
		pill = widgets.ModePillStyle("BISECT", "#1e1e2e", "#94e2d5")
	}
	if m.cfg.ActiveGitHubAccount != "" {
		if pill != "" {
			pill += "  "
		}
		pill += widgets.ModePillStyle("⬡ "+m.cfg.ActiveGitHubAccount, "#1e1e2e", "#89b4fa")
	} else if m.cfg.ActiveGitLabAccount != "" {
		if pill != "" {
			pill += "  "
		}
		pill += widgets.ModePillStyle("⬡ "+m.cfg.ActiveGitLabAccount, "#1e1e2e", "#e78284")
	}
	sb.ModePill = pill
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

// splitPanelHeight divides panelH into (tabH for files/branches, commitH for commits).
// Each panel needs at least 4 lines (border + title + 1 content + border).
func splitPanelHeight(panelH int) (tabH, commitH int) {
	const minH = 4
	tabH = panelH / 3
	if tabH < minH {
		tabH = minH
	}
	commitH = panelH - 2*tabH
	if commitH < minH {
		commitH = minH
		// Reduce tabH to fit
		tabH = (panelH - minH) / 2
		if tabH < minH {
			tabH = minH
		}
	}
	return
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
	// Save cursor positions before recreating panel structs.
	filesCur := m.files.ListCursor()
	branchCur := m.branches.ListCursor()
	commitsCur := m.commits.ListCursor()
	stashCur := m.stash.ListCursor()
	tagsCur := m.tags.ListCursor()
	worktreeCur := m.worktree.ListCursor()
	remotesCur := m.remotes.ListCursor()
	bisectCur := m.bisect.ListCursor()
	reflogCur := m.reflog.ListCursor()
	prCur := m.pr.ListCursor()
	rebaseCur := m.rebase.ListCursor()
	conflictCur := m.conflict.ListCursor()
	fileHistCur := m.fileHistory.ListCursor()
	accountsCur := m.accounts.ListCursor()
	accountsTab := m.accounts.Tab()

	layout := ComputeLayout(m.width, m.height)
	panelH, leftW, rightW := layout.LeftHeight, layout.LeftWidth, layout.RightWidth
	tabH, commitH := splitPanelHeight(panelH)

	m.files = panels.NewFilesModel(leftW, tabH)
	m.branches = panels.NewBranchModel(leftW, tabH)
	m.commits = panels.NewCommitModel(leftW, commitH)
	m.diff = panels.NewDiffModel(rightW, panelH)
	m.commitMsg = panels.NewCommitMsgModel(leftW, panelH)
	m.stash = panels.NewStashModel(m.width-2, panelH)
	m.blame = panels.NewBlameModel(m.width-2, panelH)
	m.reflog = panels.NewReflogModel(m.width-2, panelH)
	m.pr = panels.NewPRModel(m.width-2, panelH)
	m.tags = panels.NewTagsModel(m.width-2, panelH)
	m.worktree = panels.NewWorktreeModel(m.width-2, panelH)
	m.rebase = panels.NewRebaseModel(m.width-2, panelH)
	m.conflict = panels.NewConflictModel(m.width-2, panelH)
	m.fileHistory = panels.NewCommitModel(m.width-2, panelH)
	m.remotes = panels.NewRemotesModel(m.width-2, panelH)
	m.bisect = panels.NewBisectModel(m.width-2, panelH)
	m.accounts = panels.NewAccountsModel(m.width-2, panelH)
	m.accounts.LoadConfig(m.cfg)
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

	// Restore cursor positions.
	m.files.SetListCursor(filesCur)
	m.branches.SetListCursor(branchCur)
	m.commits.SetListCursor(commitsCur)
	m.stash.SetListCursor(stashCur)
	m.tags.SetListCursor(tagsCur)
	m.worktree.SetListCursor(worktreeCur)
	m.remotes.SetListCursor(remotesCur)
	m.bisect.SetListCursor(bisectCur)
	m.reflog.SetListCursor(reflogCur)
	m.pr.SetListCursor(prCur)
	m.rebase.SetListCursor(rebaseCur)
	m.conflict.SetListCursor(conflictCur)
	m.fileHistory.SetListCursor(fileHistCur)
	m.accounts.SetListCursor(accountsCur)
	if accountsTab == 1 {
		m.accounts.ToggleTab()
	}
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
		{Title: "Diff  ([/] hunks, ctrl+i: AI summary)", Bindings: []key.Binding{
			km.ToggleDiffStaged,
		}},
		{Title: "Remote", Bindings: []key.Binding{km.Push, km.Pull, km.Fetch}},
		{Title: "Branches", Bindings: []key.Binding{km.Checkout, km.NewBranch, km.RenameBranch, km.Delete, km.Merge, km.Rebase, km.OpenBrowser}},
		{Title: "Commits  (g: graph, b: blame, y: copy hash)", Bindings: []key.Binding{km.CherryPick, km.Revert, km.Reset, km.InterRebase, km.OpenBrowser}},
		{Title: "Files", Bindings: []key.Binding{km.FileHistory}},
		{Title: "Global", Bindings: []key.Binding{km.Worktree}},
		{Title: "Tags  (t: open  n: new  D: delete  P: push to remote)", Bindings: []key.Binding{km.Tags}},
		{Title: "Commit form  (tab: fields, ctrl+g: AI, ctrl+s: commit)", Bindings: []key.Binding{km.Escape}},
		{Title: "App  (ctrl+p/alt+p: palette, O: settings)", Bindings: []key.Binding{km.Settings, km.Search, km.Refresh, km.Help, km.Quit}},
	}
}

// openPRPanel is the shared helper to switch to the PR overlay.
func (m Model) openPRPanel() (tea.Model, tea.Cmd) {
	m.prevFocused = m.focused
	m.focused = PanelPR
	m.mode = ModePR
	return m, m.loadPRs()
}

// handleAccountsKey handles input while the accounts panel is focused.
func (m Model) handleAccountsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == ModeAddAccount {
		switch msg.String() {
		case "esc":
			m.mode = ModeAccounts
			m.accounts.HideModal()
		case "enter":
			done, name, token, host := m.accounts.AdvanceAddModal()
			if done {
				if name == "" {
					m.mode = ModeAccounts
					m.setStatus("Account name cannot be empty", true)
					return m, nil
				}
				m.mode = ModeAccounts
				m.accounts.AddAccount(&m.cfg, name, token, host)
				cfg := m.cfg
				return m, func() tea.Msg {
					if err := config.Save(cfg); err != nil {
						return StatusMsg{Text: "Account added (config write failed: " + err.Error() + ")", IsErr: true}
					}
					return StatusMsg{Text: "Account '" + name + "' added ✓", IsErr: false}
				}
			}
		default:
			return m, m.accounts.UpdateModalInput(msg)
		}
		return m, nil
	}

	switch msg.String() {
	case "tab":
		m.accounts.ToggleTab()
	case "j", "down":
		m.accounts.MoveDown()
	case "k", "up":
		m.accounts.MoveUp()
	case "n":
		m.accounts.ShowAddModal(m.accounts.CurrentForgeType())
		m.mode = ModeAddAccount
	case "D":
		if a := m.accounts.CurrentAccount(); a != nil {
			acctName := a.Name
			ft := m.accounts.CurrentForgeType()
			m.confirmModal = widgets.NewConfirmModal(
				"Delete Account?",
				"Delete '"+acctName+"' from sugi config?\n\nThis does not revoke the token.",
			)
			m.confirmModal.Show()
			m.mode = ModeConfirm
			m.confirmAction = func() tea.Cmd {
				return func() tea.Msg {
					return AccountDeleteMsg{Name: acctName, ForgeType: ft}
				}
			}
		}
	case "enter":
		if a := m.accounts.CurrentAccount(); a != nil {
			return m.switchAccount(a.Name, m.accounts.CurrentForgeType())
		}
	case "esc":
		m.mode = ModeNormal
		m.focused = m.prevFocused
	}
	return m, nil
}

// switchAccount activates a named account, recreates the forge client, and saves config.
func (m Model) switchAccount(name, forgeType string) (tea.Model, tea.Cmd) {
	switch forgeType {
	case "github":
		m.cfg.ActiveGitHubAccount = name
		m.accounts.SetActive(name, "github")
		if m.forge != nil {
			info := m.forge.ForgeInfo()
			if info.Type == forge.ForgeGitHub {
				m.forge = forge.NewGitHubClient(info, m.cfg.ActiveGitHubToken())
			}
		}
	case "gitlab":
		m.cfg.ActiveGitLabAccount = name
		m.accounts.SetActive(name, "gitlab")
		if m.forge != nil {
			info := m.forge.ForgeInfo()
			if info.Type == forge.ForgeGitLab {
				m.forge = forge.NewGitLabClient(info, m.cfg.ActiveGitLabToken())
			}
		}
	}
	cfg := m.cfg
	return m, func() tea.Msg {
		if err := config.Save(cfg); err != nil {
			return StatusMsg{Text: "Switched to " + name + " (config write failed: " + err.Error() + ")", IsErr: true}
		}
		return StatusMsg{Text: "Switched to account: " + name + " ✓", IsErr: false}
	}
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
	case "open_worktrees":
		m.prevFocused = m.focused
		m.focused = PanelWorktree
		m.mode = ModeWorktree
		return m, m.loadWorktrees()
	case "open_remotes":
		m.prevFocused = m.focused
		m.focused = PanelRemotes
		m.mode = ModeRemotes
		return m, m.loadRemotes()
	case "open_bisect":
		m.prevFocused = m.focused
		m.focused = PanelBisect
		m.mode = ModeBisect
		return m, m.loadBisectStatus()
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
		{ID: "open_worktrees", Label: "Open worktrees panel", Keys: "W", Category: "Git"},
		{ID: "open_remotes", Label: "Open remotes panel", Keys: "E", Category: "Git"},
		{ID: "open_bisect", Label: "Open git bisect panel", Keys: "B", Category: "Git"},
		{ID: "toggle_graph", Label: "Toggle commit graph", Keys: "g", Category: "Commits"},
		{ID: "toggle_staged", Label: "Toggle staged diff view", Keys: "S", Category: "Diff"},
		{ID: "ai_commit", Label: "AI: generate commit message", Keys: "ctrl+g", Category: "AI"},
		{ID: "ai_diff_summary", Label: "AI: summarise current diff", Keys: "ctrl+i", Category: "AI"},
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
