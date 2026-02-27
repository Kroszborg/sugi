# Changelog

All notable changes to sugi are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
sugi uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added
- **Bisect panel** (`B`) — interactive `git bisect` session with good/bad marking, step counter, log view
- **Remotes panel** (`E`) — list, add (two-step modal), remove, rename, fetch remotes
- **Worktrees panel** (`W`) — list, add, remove git worktrees
- **Interactive rebase** (`i` from commits) — reorder, squash, fixup, drop, reword commits; color-coded actions
- **Merge conflict resolver** — opens automatically on conflicted files; choose ours/theirs per block
- **File history** (`L`) — log of commits that touched the currently selected file
- **Merge branch** (`m` in branches) — merge selected branch into current with abort support
- **Rebase branch** (`r` in branches) — rebase current branch onto selected; abort/continue aware
- **Rename branch** (`R` in branches) — in-place modal pre-filled with current name
- **Revert commit** (`v` in commits) — `git revert --no-edit` with confirmation
- **Reset HEAD** (`X` in commits) — soft / mixed / hard, prompted inline
- **Open in browser** (`o`) — opens branch or commit on GitHub / GitLab
- **Multi-select files** (`ctrl+space`) — bulk stage/unstage/discard
- **Panel icons** — `▸ FILES`, `⎇ BRANCHES`, `● COMMITS`, `≋ DIFF`
- **MERGE / REBASE in-progress badges** in header bar
- **BISECT mode pill** in status bar during active bisect session
- **Toast notifications** — auto-expiring 3 s success/error toasts
- **Scrolllist gutter cursor** — `▶` indicator replaces fragile background highlight

### Changed
- Status bar shows `·` separators and context-aware mode pill
- Help overlay is scrollable (`j`/`k`)
- `r` key scoped per panel (refresh globally, rebase in branches panel)

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
