package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds all keybindings for sugi.
// This single struct drives both input handling and the help overlay.
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding

	// Panel focus
	FocusFiles    key.Binding
	FocusBranches key.Binding
	FocusCommits  key.Binding
	FocusDiff     key.Binding
	NextPanel     key.Binding
	PrevPanel     key.Binding

	// Git actions
	Stage     key.Binding
	Unstage   key.Binding
	StageAll  key.Binding
	Discard   key.Binding
	Commit    key.Binding
	Push      key.Binding
	Pull      key.Binding
	Fetch     key.Binding
	Checkout  key.Binding
	NewBranch key.Binding
	Delete    key.Binding

	// View toggles
	ToggleSideBySide key.Binding
	ToggleDiffStaged key.Binding

	// App
	Palette  key.Binding
	Settings key.Binding
	Help     key.Binding
	Quit     key.Binding
	Refresh  key.Binding
	Search   key.Binding
	Escape   key.Binding
	Confirm  key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+b"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+f"), key.WithHelp("pgdn", "page down")),
		Top:      key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
		Bottom:   key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),

		FocusFiles:    key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "files")),
		FocusBranches: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "branches")),
		FocusCommits:  key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "commits")),
		FocusDiff:     key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "diff")),
		NextPanel:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		PrevPanel:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),

		Stage:     key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "stage/unstage")),
		Unstage:   key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "unstage")),
		StageAll:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "stage all")),
		Discard:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "discard")),
		Commit:    key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "commit")),
		Push:      key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "push")),
		Pull:      key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pull")),
		Fetch:     key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fetch")),
		Checkout:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "checkout/select")),
		NewBranch: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new branch")),
		Delete:    key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "delete")),

		ToggleSideBySide: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "side-by-side diff")),
		ToggleDiffStaged: key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "toggle staged diff")),

		Palette:  key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "command palette")),
		Settings: key.NewBinding(key.WithKeys("O"), key.WithHelp("O", "settings")),
		Help:     key.NewBinding(key.WithKeys("?", "shift+/"), key.WithHelp("?", "help")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Search:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Escape:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back/cancel")),
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	}
}
