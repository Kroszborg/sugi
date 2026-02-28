package panels

import (
	"fmt"
	"strings"

	"github.com/Kroszborg/sugi/internal/config"
	"github.com/Kroszborg/sugi/internal/ui/widgets"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AccountsModel manages GitHub/GitLab named account tokens.
type AccountsModel struct {
	github []config.AccountEntry
	gitlab []config.AccountEntry
	tab    int // 0 = GitHub, 1 = GitLab
	list   widgets.ScrollList
	Width  int
	Height int

	// Add-account modal state
	addModal     widgets.Modal
	showModal    bool
	addStep      int    // 0 = name, 1 = token, 2 = host
	addName      string // name buffer across steps
	addToken     string // token buffer across steps
	addForgeType string // "github" or "gitlab" — set when modal opens

	// Active account names (mirrors cfg fields for display)
	ActiveGitHub string
	ActiveGitLab string

	// styles
	titleStyle  lipgloss.Style
	tabStyle    lipgloss.Style
	activeTab   lipgloss.Style
	hintStyle   lipgloss.Style
	activeStyle lipgloss.Style
	nameStyle   lipgloss.Style
	tokenStyle  lipgloss.Style
	hostStyle   lipgloss.Style
	pillStyle   lipgloss.Style
}

// NewAccountsModel creates an AccountsModel sized to the overlay dimensions.
func NewAccountsModel(width, height int) AccountsModel {
	m := AccountsModel{
		Width:  width,
		Height: height,
		list:   widgets.NewScrollList(height-6, width-4),
	}
	m.titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true)
	m.tabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
	m.activeTab = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true).Underline(true)
	m.hintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70"))
	m.activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true)
	m.nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Bold(true)
	m.tokenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7"))
	m.hostStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
	m.pillStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#a6e3a1")).
		Foreground(lipgloss.Color("#1e1e2e")).
		Bold(true).
		Padding(0, 1)
	return m
}

// LoadConfig reloads account lists and active selections from config.
func (m *AccountsModel) LoadConfig(cfg config.Config) {
	m.github = cfg.GitHubAccounts
	m.gitlab = cfg.GitLabAccounts
	m.ActiveGitHub = cfg.ActiveGitHubAccount
	m.ActiveGitLab = cfg.ActiveGitLabAccount
	m.refreshList()
}

// SetActive updates the displayed active account for the given forge type.
func (m *AccountsModel) SetActive(name, forgeType string) {
	if forgeType == "github" {
		m.ActiveGitHub = name
	} else {
		m.ActiveGitLab = name
	}
	m.refreshList()
}

// Tab returns the current tab index (0=GitHub, 1=GitLab).
func (m *AccountsModel) Tab() int { return m.tab }

// ToggleTab switches between the GitHub and GitLab tabs.
func (m *AccountsModel) ToggleTab() {
	m.tab = 1 - m.tab
	m.list.Cursor = 0
	m.refreshList()
}

// MoveUp moves the list cursor up.
func (m *AccountsModel) MoveUp() { m.list.MoveUp() }

// MoveDown moves the list cursor down.
func (m *AccountsModel) MoveDown() { m.list.MoveDown() }

// ListCursor returns the current scroll list cursor position.
func (m *AccountsModel) ListCursor() int { return m.list.Cursor }

// SetListCursor sets the scroll list cursor position.
func (m *AccountsModel) SetListCursor(n int) { m.list.Cursor = n }

// CurrentAccount returns the selected account or nil if the list is empty.
func (m *AccountsModel) CurrentAccount() *config.AccountEntry {
	switch m.tab {
	case 0:
		if len(m.github) == 0 || m.list.Cursor >= len(m.github) {
			return nil
		}
		return &m.github[m.list.Cursor]
	case 1:
		if len(m.gitlab) == 0 || m.list.Cursor >= len(m.gitlab) {
			return nil
		}
		return &m.gitlab[m.list.Cursor]
	}
	return nil
}

// CurrentForgeType returns "github" or "gitlab" depending on the active tab.
func (m *AccountsModel) CurrentForgeType() string {
	if m.tab == 0 {
		return "github"
	}
	return "gitlab"
}

// ShowAddModal opens the 3-step add-account modal at step 0 (name entry).
func (m *AccountsModel) ShowAddModal(forgeType string) {
	m.addStep = 0
	m.addName = ""
	m.addToken = ""
	m.addForgeType = forgeType
	mod := widgets.NewInputModal("Add Account — Step 1 of 3: Name", "e.g. personal, work")
	mod.Body = "A short nickname to identify this account.\nExamples: personal  work  client"
	mod.Show()
	m.addModal = mod
	m.showModal = true
}

// AdvanceAddModal accepts the current step's value and moves to the next.
// Returns (done=true, name, token, host) when all three steps are complete.
func (m *AccountsModel) AdvanceAddModal() (done bool, name, token, host string) {
	val := m.addModal.Input.Value()
	switch m.addStep {
	case 0:
		m.addName = val
		m.addStep = 1
		mod := widgets.NewInputModal("Add Account — Step 2 of 3: Token", tokenPlaceholder(m.addForgeType))
		mod.Body = tokenHint(m.addForgeType)
		mod.Show()
		m.addModal = mod
		return false, "", "", ""
	case 1:
		m.addToken = val
		m.addStep = 2
		mod := widgets.NewInputModal("Add Account — Step 3 of 3: Host (optional)", "e.g. github.mycompany.com")
		mod.Body = "Leave blank for github.com / gitlab.com\nOnly needed for GitHub Enterprise or self-hosted GitLab\nExample: github.mycompany.com"
		mod.Show()
		m.addModal = mod
		return false, "", "", ""
	case 2:
		h := val
		m.HideModal()
		return true, m.addName, m.addToken, h
	}
	return false, "", "", ""
}

func tokenPlaceholder(forgeType string) string {
	if forgeType == "gitlab" {
		return "glpat-..."
	}
	return "ghp_... or github_pat_..."
}

func tokenHint(forgeType string) string {
	if forgeType == "gitlab" {
		return "Create at: gitlab.com/-/profile/personal_access_tokens\nRequired scope: api\nToken starts with: glpat-"
	}
	return "Create at: github.com/settings/tokens\nRequired scopes: repo + read:user\nToken starts with: ghp_  or  github_pat_"
}

// IsModalVisible reports whether the add-account modal is open.
func (m *AccountsModel) IsModalVisible() bool { return m.showModal }

// HideModal dismisses the add-account modal.
func (m *AccountsModel) HideModal() {
	m.addModal.Hide()
	m.showModal = false
	m.addStep = 0
	m.addName = ""
	m.addToken = ""
}

// UpdateModalInput forwards a key event to the active modal's text input.
func (m *AccountsModel) UpdateModalInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.addModal.Input, cmd = m.addModal.Input.Update(msg)
	return cmd
}

// ModalInput returns the current value of the modal text input.
func (m *AccountsModel) ModalInput() string {
	return m.addModal.Input.Value()
}

// ModalView renders the active add-account modal.
func (m *AccountsModel) ModalView() string {
	return m.addModal.View()
}

// AddAccount appends a new account to the current tab's list and updates cfg.
func (m *AccountsModel) AddAccount(cfg *config.Config, name, token, host string) {
	entry := config.AccountEntry{Name: name, Token: token, Host: host}
	if m.tab == 0 {
		cfg.GitHubAccounts = append(cfg.GitHubAccounts, entry)
		m.github = cfg.GitHubAccounts
	} else {
		cfg.GitLabAccounts = append(cfg.GitLabAccounts, entry)
		m.gitlab = cfg.GitLabAccounts
	}
	m.refreshList()
}

// DeleteByName removes the named account from the given forge type list.
func (m *AccountsModel) DeleteByName(cfg *config.Config, name, forgeType string) {
	switch forgeType {
	case "github":
		for i, a := range m.github {
			if a.Name == name {
				m.github = append(m.github[:i], m.github[i+1:]...)
				cfg.GitHubAccounts = m.github
				if cfg.ActiveGitHubAccount == name {
					cfg.ActiveGitHubAccount = ""
					m.ActiveGitHub = ""
				}
				if m.list.Cursor >= len(m.github) && m.list.Cursor > 0 {
					m.list.Cursor--
				}
				break
			}
		}
	case "gitlab":
		for i, a := range m.gitlab {
			if a.Name == name {
				m.gitlab = append(m.gitlab[:i], m.gitlab[i+1:]...)
				cfg.GitLabAccounts = m.gitlab
				if cfg.ActiveGitLabAccount == name {
					cfg.ActiveGitLabAccount = ""
					m.ActiveGitLab = ""
				}
				if m.list.Cursor >= len(m.gitlab) && m.list.Cursor > 0 {
					m.list.Cursor--
				}
				break
			}
		}
	}
	m.refreshList()
}

// View renders the accounts panel content (placed inside renderOverlay border by app.go).
func (m *AccountsModel) View() string {
	var sb strings.Builder

	// Tab header
	ghLabel := " GitHub "
	glLabel := " GitLab "
	if m.tab == 0 {
		sb.WriteString(m.activeTab.Render(ghLabel) + "  " + m.tabStyle.Render(glLabel))
	} else {
		sb.WriteString(m.tabStyle.Render(ghLabel) + "  " + m.activeTab.Render(glLabel))
	}
	sb.WriteString("\n\n")

	// Account list or empty state
	var accounts []config.AccountEntry
	if m.tab == 0 {
		accounts = m.github
	} else {
		accounts = m.gitlab
	}

	if len(accounts) == 0 {
		sb.WriteString(m.hintStyle.Render("  No accounts configured — press n to add one"))
	} else {
		sb.WriteString(m.list.View())
	}

	return sb.String()
}

// refreshList rebuilds the scrolllist items from the current tab's accounts.
func (m *AccountsModel) refreshList() {
	var items []string
	switch m.tab {
	case 0:
		for _, a := range m.github {
			items = append(items, m.renderEntry(a, a.Name == m.ActiveGitHub))
		}
	case 1:
		for _, a := range m.gitlab {
			items = append(items, m.renderEntry(a, a.Name == m.ActiveGitLab))
		}
	}
	m.list.SetItems(items)
}

func (m *AccountsModel) renderEntry(a config.AccountEntry, isActive bool) string {
	prefix := "  "
	if isActive {
		prefix = m.activeStyle.Render("▶ ")
	}

	name := m.nameStyle.Render(a.Name)

	// Mask token: show last 4 chars only
	masked := "●●●●"
	if len(a.Token) > 4 {
		masked = "●●●●" + a.Token[len(a.Token)-4:]
	}
	tok := m.tokenStyle.Render(masked)

	host := ""
	if a.Host != "" {
		host = "  " + m.hostStyle.Render(a.Host)
	}

	active := ""
	if isActive {
		active = "  " + m.pillStyle.Render("active")
	}

	return fmt.Sprintf("%s%s  %s%s%s", prefix, name, tok, host, active)
}
