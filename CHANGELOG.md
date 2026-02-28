# Changelog

All notable changes to sugi are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
sugi uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

---

## [0.3.4] — 2026-02-28

### Added
- **Accounts panel** (`A`) — manage multiple GitHub/GitLab tokens with named accounts; switch active account in-app; persisted to config
- **Multi-account support** — `active_github_account` / `active_gitlab_account` in config; status bar pill shows active account name
- **Command palette** (`ctrl+p` / `alt+p`) — fuzzy-search all commands; `alt+p` works in VS Code/Cursor terminals
- **Bisect panel** (`B`) — interactive `git bisect` session with good/bad marking, step counter, log view
- **Remotes panel** (`E`) — list, add (two-step modal), remove, rename, fetch remotes
- **Worktrees panel** (`W`) — list, add, remove git worktrees
- **Interactive rebase** (`i` from commits) — reorder, squash, fixup, drop, reword commits; color-coded actions
- **Merge conflict resolver** — opens automatically on conflicted files; choose ours/theirs per block
- **File history** (`L`) — log of commits that touched the currently selected file

### Changed
- **Deep Noir color palette** — Linear/Vercel-inspired dark theme; electric violet `#7c6dfa` as single primary accent (replaces catppuccin sky)
- **Diff panel**: full-width background tinting on added (green) / removed (red) lines; GitHub-style
- **Diff scrollbar**: minimal `▌` thumb indicator, invisible track — content stays within panel bounds (no window expansion)
- **ScrollList scrollbar**: same thumb-only indicator on all panels
- **Focused panel titles**: underline + accent color when active; clearer focus state
- **Cursor row**: violet-tinted background, more readable in all panels
- **Confirm modals**: `[y] confirm  [esc] cancel` — one-click cancel, no redundant `[n] No`
- Amend shortcut changed: `ctrl+a` (was `A` — conflict with accounts panel)
- AI diff summary shortcut changed: `ctrl+i` (was `A` in diff panel)

### Fixed
- Diff panel no longer expands terminal window when diff is longer than panel height
- ScrollList height calculation: `height-3` (was `height-4`, causing 1 wasted row per panel)
- Input modals (add account, add remote, new tag) no longer hijack global keys while typing
- `sugi version` now works; unknown subcommands show proper error (removed `SilenceErrors`)
- npm install: redirect handling fixed, `go install` fallback uses `@latest` with version ldflags
- npm package name corrected to `@kroszborg/sugi` throughout docs and goreleaser config

---

## [0.1.1] — 2025-01-xx

### Added
- npm package for cross-platform install (`npm install -g sugi`)
- Settings panel (`O`) — edit config in-app, saved to disk immediately
- AI commit generation via Groq (`ctrl+g` / `alt+g`)
- AI diff summary (`A` in diff panel)
- Stash panel (`z`) with diff preview
- Blame panel (`b`)
- Reflog panel with undo
- Commit graph toggle (`g`)
- Hunk-level staging (`space` / `u` in diff panel)
- Commit form with subject + body, 72-char counter
- Command palette (`ctrl+p`)
- Help overlay (`?`, auto-generated from keymap)
- Cherry-pick (`C` in commits)

### Fixed
- Config file created with `0o600` permissions (owner-only read/write)

---

## [0.1.0] — 2025-01-xx

### Added
- Initial release
- Files, Branches, Commits, Diff panels
- GitHub / GitLab PR panel with CI badges
- Stage, unstage, discard, commit, push, pull, fetch
- Branch create/checkout/delete
- Tag management
