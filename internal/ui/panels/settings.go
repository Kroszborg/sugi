package panels

import (
	"strings"

	"github.com/Kroszborg/sugi/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settingKind int

const (
	settingKindText   settingKind = iota // plain text
	settingKindMasked                    // API keys — show partial mask
	settingKindBool                      // toggle with space/enter
)

type settingField struct {
	Label string
	Key   string
	Value string
	Kind  settingKind
	Hint  string
}

// SettingsModel is a full-screen settings overlay with inline editing.
type SettingsModel struct {
	fields  []settingField
	cursor  int
	editing bool
	input   textinput.Model
	width   int
	height  int

	overlayStyle  lipgloss.Style
	titleStyle    lipgloss.Style
	sectionStyle  lipgloss.Style
	labelStyle    lipgloss.Style
	valueStyle    lipgloss.Style
	maskedStyle   lipgloss.Style
	unsetStyle    lipgloss.Style
	cursorStyle   lipgloss.Style
	enabledStyle  lipgloss.Style
	disabledStyle lipgloss.Style
	hintStyle     lipgloss.Style
	editStyle     lipgloss.Style
	footerStyle   lipgloss.Style
}

// NewSettingsModel creates a SettingsModel from config.
func NewSettingsModel(cfg config.Config, width, height int) SettingsModel {
	inp := textinput.New()
	inp.CharLimit = 256
	m := SettingsModel{input: inp}
	m.buildStyles(width, height)
	m.LoadConfig(cfg)
	return m
}

func (m *SettingsModel) buildStyles(width, height int) {
	w := width - 8
	if w > 74 {
		w = 74
	}
	if w < 40 {
		w = 40
	}
	h := height - 6
	if h < 14 {
		h = 14
	}
	m.width = w
	m.height = h

	m.overlayStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4d9de0")).
		Padding(1, 2).
		Width(w).Height(h)
	m.titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4d9de0")).Bold(true)
	m.sectionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e8835c")).Bold(true)
	m.labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7878a0")).Width(22)
	m.valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee"))
	m.maskedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a87efb"))
	m.unsetStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#252538"))
	m.cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4d9de0")).Bold(true)
	m.enabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3ecf8e")).Bold(true)
	m.disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c"))
	m.hintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c"))
	m.editStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d4a017")).Bold(true)
	m.footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3d3d5c"))
}

// LoadConfig populates fields from a Config value.
func (m *SettingsModel) LoadConfig(cfg config.Config) {
	groqModel := cfg.GroqModel
	if groqModel == "" {
		groqModel = "llama-3.1-8b-instant"
	}
	boolStr := func(b bool) string {
		if b {
			return "true"
		}
		return "false"
	}
	m.fields = []settingField{
		// AI
		{Label: "Groq API Key", Key: "groq_api_key", Value: cfg.GroqAPIKey, Kind: settingKindMasked,
			Hint: "free key at console.groq.com — or set GROQ_API_KEY env var"},
		{Label: "Groq Model", Key: "groq_model", Value: groqModel, Kind: settingKindText,
			Hint: "llama-3.1-8b-instant (fast, free)   llama-3.3-70b-versatile (smarter)"},
		// Forge
		{Label: "GitHub Token", Key: "github_token", Value: cfg.GitHubToken, Kind: settingKindMasked,
			Hint: "or set GITHUB_TOKEN / GH_TOKEN env var (sugi reads gh CLI automatically)"},
		{Label: "GitLab Token", Key: "gitlab_token", Value: cfg.GitLabToken, Kind: settingKindMasked,
			Hint: "or set GITLAB_TOKEN env var"},
		{Label: "GitLab Host", Key: "gitlab_host", Value: cfg.GitLabHost, Kind: settingKindText,
			Hint: "self-hosted only — e.g. https://gitlab.company.com"},
		// UI
		{Label: "Mouse Enabled", Key: "mouse_enabled", Value: boolStr(cfg.MouseEnabled), Kind: settingKindBool,
			Hint: "space or enter to toggle"},
		{Label: "Show Graph", Key: "show_graph", Value: boolStr(cfg.ShowGraph), Kind: settingKindBool,
			Hint: "space or enter to toggle   (also: g in commits panel)"},
	}
}

// BuildConfig converts current field values back to a Config.
func (m SettingsModel) BuildConfig(base config.Config) config.Config {
	cfg := base
	for _, f := range m.fields {
		switch f.Key {
		case "groq_api_key":
			cfg.GroqAPIKey = f.Value
		case "groq_model":
			cfg.GroqModel = f.Value
		case "github_token":
			cfg.GitHubToken = f.Value
		case "gitlab_token":
			cfg.GitLabToken = f.Value
		case "gitlab_host":
			cfg.GitLabHost = f.Value
		case "mouse_enabled":
			cfg.MouseEnabled = f.Value == "true"
		case "show_graph":
			cfg.ShowGraph = f.Value == "true"
		}
	}
	return cfg
}

// Resize updates the overlay dimensions (called on terminal resize).
func (m *SettingsModel) Resize(width, height int) {
	m.buildStyles(width, height)
}

func (m *SettingsModel) MoveUp() {
	if m.editing {
		return
	}
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *SettingsModel) MoveDown() {
	if m.editing {
		return
	}
	if m.cursor < len(m.fields)-1 {
		m.cursor++
	}
}

// Toggle flips a bool field at the cursor.
func (m *SettingsModel) Toggle() {
	if m.cursor >= len(m.fields) {
		return
	}
	if m.fields[m.cursor].Kind != settingKindBool {
		return
	}
	if m.fields[m.cursor].Value == "true" {
		m.fields[m.cursor].Value = "false"
	} else {
		m.fields[m.cursor].Value = "true"
	}
}

// StartEdit begins editing the current field (or toggles bool fields).
func (m *SettingsModel) StartEdit() {
	if m.cursor >= len(m.fields) {
		return
	}
	f := m.fields[m.cursor]
	if f.Kind == settingKindBool {
		m.Toggle()
		return
	}
	m.editing = true
	m.input.SetValue(f.Value)
	m.input.Focus()
	m.input.CursorEnd()
}

// ConfirmEdit saves the text input value and exits edit mode.
func (m *SettingsModel) ConfirmEdit() {
	if !m.editing {
		return
	}
	m.fields[m.cursor].Value = m.input.Value()
	m.editing = false
	m.input.Blur()
}

// CancelEdit discards in-progress edits.
func (m *SettingsModel) CancelEdit() {
	m.editing = false
	m.input.Blur()
}

// IsEditing reports whether a text field is being edited.
func (m SettingsModel) IsEditing() bool { return m.editing }

// UpdateInput forwards a key event to the text input and returns the updated model + cmd.
func (m SettingsModel) UpdateInput(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
	if !m.editing {
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m SettingsModel) renderValue(idx int, focused bool) string {
	f := m.fields[idx]
	if focused && m.editing {
		return m.editStyle.Render("✎ ") + m.input.View()
	}
	if f.Value == "" {
		return m.unsetStyle.Render("(not set)")
	}
	switch f.Kind {
	case settingKindBool:
		if f.Value == "true" {
			return m.enabledStyle.Render("✓ enabled")
		}
		return m.disabledStyle.Render("✗ disabled")
	case settingKindMasked:
		n := len(f.Value)
		if n == 0 {
			return m.unsetStyle.Render("(not set)")
		}
		visible := 4
		if n <= visible {
			visible = 0
		}
		dots := n - visible
		if dots > 14 {
			dots = 14
		}
		return m.maskedStyle.Render(f.Value[:visible] + strings.Repeat("•", dots))
	default:
		v := f.Value
		maxLen := m.width - 30
		if maxLen < 20 {
			maxLen = 20
		}
		if len(v) > maxLen {
			v = v[:maxLen-1] + "…"
		}
		return m.valueStyle.Render(v)
	}
}

// View renders the settings overlay.
func (m SettingsModel) View() string {
	if m.width == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(m.titleStyle.Render("⚙  Settings") + "\n")
	sb.WriteString(m.hintStyle.Render("  Config: ~/.config/sugi/config.json") + "\n\n")

	type section struct {
		name   string
		fields []int
	}
	sections := []section{
		{"AI", []int{0, 1}},
		{"Forge", []int{2, 3, 4}},
		{"UI", []int{5, 6}},
	}

	for _, sec := range sections {
		divider := strings.Repeat("─", m.width-len(sec.name)-5)
		sb.WriteString(m.sectionStyle.Render(sec.name) + " " + m.hintStyle.Render(divider) + "\n\n")
		for _, idx := range sec.fields {
			if idx >= len(m.fields) {
				continue
			}
			focused := idx == m.cursor
			label := m.labelStyle.Render(m.fields[idx].Label)
			value := m.renderValue(idx, focused)

			if focused {
				prefix := m.cursorStyle.Render("▶")
				sb.WriteString(" " + prefix + " " + label + " " + value + "\n")
				sb.WriteString(m.hintStyle.Render("     ↳ "+m.fields[idx].Hint) + "\n\n")
			} else {
				sb.WriteString("   " + label + " " + value + "\n\n")
			}
		}
	}

	editHint := "enter edit  space toggle"
	if m.editing {
		editHint = "enter confirm  esc cancel edit"
	}
	sb.WriteString(m.footerStyle.Render("  ↑↓/jk " + editHint + "  ctrl+s save  esc back"))
	sb.WriteString("\n" + m.footerStyle.Render("  A  Manage accounts →"))

	content := sb.String()
	// Pre-truncate content to available height (border=2, padding=2 → available = h-4)
	avail := m.height - 4
	if avail > 0 {
		lines := strings.Split(content, "\n")
		if len(lines) > avail {
			lines = lines[:avail]
			content = strings.Join(lines, "\n")
		}
	}

	return m.overlayStyle.Render(content)
}
