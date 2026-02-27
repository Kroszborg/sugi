# sugi 杉

> A terminal UI git client — GitHub/GitLab PRs, AI commit messages, interactive rebase, bisect, worktrees, and more.

```
  sugi (杉) — cedar tree. Grows fast, stands tall, shaped with precision.
```

[![CI](https://github.com/Kroszborg/sugi/actions/workflows/ci.yml/badge.svg)](https://github.com/Kroszborg/sugi/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Kroszborg/sugi)](https://goreportcard.com/report/github.com/Kroszborg/sugi)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Screenshots

```
 ⬡ sugi  ▸ FILES  ⎇ BRANCHES  ● COMMITS  ≋ DIFF
┌──────────────┬────────────────┬──────────────────────────┐
│ ▶ M src/     │ ▶ main         │ ▶ feat: add bisect UI    │
│ · M go.sum   │ · feature/x    │ · fix: scrolllist cursor │
│              │ · bugfix/y     │ · chore: update deps     │
└──────────────┴────────────────┴──────────────────────────┘
```

---

## Features

### Core panels
| Panel | Key | What it does |
|-------|-----|--------------|
| **Files** | `1` | Stage/unstage/discard individual files or hunks; multi-select with `ctrl+space` |
| **Branches** | `2` | Checkout, create, rename, delete; merge, rebase, open in browser |
| **Commits** | `3` | Full log with ASCII graph, cherry-pick, revert, reset, interactive rebase, file history |
| **Diff** | `4` | Unified diff, hunk navigation, stage/unstage hunks, AI summary, side-by-side toggle |

### Extra panels (overlays)
| Panel | Key | What it does |
|-------|-----|--------------|
| **PR / MR** | `p` | GitHub & GitLab pull requests with CI badges, review status |
| **Stash** | `z` | List, apply, pop, drop stashes with diff preview |
| **Blame** | `b` | File blame — author, date, hash per line |
| **Reflog** | — | Full reflog with undo capability |
| **Worktrees** | `W` | List, add, remove git worktrees |
| **Remotes** | `E` | List, add, remove, rename, fetch remotes |
| **Bisect** | `B` | Interactive git bisect — mark good/bad, view log |
| **Interactive Rebase** | `i` | Reorder/squash/fixup/drop/reword commits visually |
| **Conflict Resolver** | auto | Opens on conflicted files — pick ours/theirs per block |
| **File History** | `L` | Log of commits touching the selected file |
| **Command Palette** | `ctrl+p` | Fuzzy search all actions |
| **Help** | `?` | Scrollable keybinding reference |
| **Settings** | `O` | Edit config in-app, saved instantly |

### AI integration
- **`ctrl+g`** — generate a commit message from staged diff (Groq, free)
- **`A`** — AI-summarise the current diff hunk
- Uses `llama-3.1-8b-instant` by default (fast, free tier)

### Git operations
Stage, unstage, discard, commit, push, pull, fetch, merge, rebase, reset (soft/mixed/hard), revert, cherry-pick, create/delete/rename branch, add/delete tag, worktree management, bisect, interactive rebase, conflict resolution.

---

## Install

```sh
# npm (macOS, Linux, Windows)
npm install -g sugi

# Homebrew (macOS / Linux)
brew install Kroszborg/tap/sugi

# go install
go install github.com/Kroszborg/sugi/cmd/sugi@latest
```

---

## Usage

```sh
sugi              # open in current directory
sugi /path/repo   # open a specific repo
sugi version      # print version
```

---

## Key bindings

### Navigation
| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | cycle panels |
| `1` `2` `3` `4` | jump to Files / Branches / Commits / Diff |
| `↑↓` / `j` `k` | move up/down |
| `←→` / `h` `l` | scroll left/right (diff) |
| `g` / `G` | jump to top / bottom |
| `pgup` / `pgdn` | page up / down |
| `/` | search / filter |
| `esc` | back / cancel |

### Files panel
| Key | Action |
|-----|--------|
| `space` | stage / unstage |
| `ctrl+space` | multi-select toggle |
| `a` | stage all |
| `u` | unstage |
| `d` | discard changes |
| `L` | file history (commits touching this file) |
| `enter` | open conflict resolver (on conflicted files) |

### Branches panel
| Key | Action |
|-----|--------|
| `enter` | checkout branch |
| `n` | new branch |
| `R` | rename branch |
| `D` | delete branch |
| `m` | merge branch into current |
| `r` | rebase current onto branch |
| `o` | open branch on GitHub / GitLab |
| `E` | remotes panel |

### Commits panel
| Key | Action |
|-----|--------|
| `enter` | show diff |
| `C` | cherry-pick commit |
| `v` | revert commit |
| `X` | reset HEAD (soft / mixed / hard) |
| `i` | interactive rebase from this commit |
| `o` | open commit on GitHub / GitLab |
| `g` | toggle commit graph |
| `b` | blame file at this commit |
| `L` | file history |

### Diff panel
| Key | Action |
|-----|--------|
| `[` / `]` | previous / next hunk |
| `space` | stage hunk |
| `u` | unstage hunk |
| `s` | side-by-side diff toggle |
| `S` | toggle staged diff |
| `A` | AI-summarise diff |

### Global
| Key | Action |
|-----|--------|
| `c` | commit form |
| `ctrl+g` | AI-generate commit message |
| `P` | push |
| `p` | pull / open PRs |
| `f` | fetch |
| `z` | stash panel |
| `W` | worktrees panel |
| `B` | bisect panel |
| `ctrl+p` | command palette |
| `O` | settings |
| `?` | help overlay |
| `q` / `ctrl+c` | quit |

---

## AI setup (Groq — free)

1. Sign up at **[console.groq.com](https://console.groq.com)** (free, no credit card needed)
2. Create an API key
3. Add it to `~/.config/sugi/config.json`:

```json
{
  "groq_api_key": "gsk_..."
}
```

Or set the environment variable:
```sh
export GROQ_API_KEY=gsk_...
```

sugi uses `llama-3.1-8b-instant` by default. To use a larger model:
```json
{ "groq_model": "llama-3.3-70b-versatile" }
```

---

## GitHub / GitLab integration

sugi auto-detects the forge from your `origin` remote URL.

**GitHub:** set `GITHUB_TOKEN`, or sugi reads from the `gh` CLI automatically.

**GitLab:** set `GITLAB_TOKEN`.

Or configure in `~/.config/sugi/config.json`:
```json
{
  "github_token": "ghp_...",
  "gitlab_token": "glpat-...",
  "gitlab_host": "https://gitlab.company.com"
}
```

---

## Full config reference

`~/.config/sugi/config.json` — all fields optional, auto-created on first run:

```json
{
  "groq_api_key": "",
  "groq_model": "llama-3.1-8b-instant",
  "github_token": "",
  "gitlab_token": "",
  "gitlab_host": "",
  "mouse_enabled": true,
  "show_graph": false
}
```

---

## Build from source

```sh
git clone https://github.com/Kroszborg/sugi
cd sugi
go build -o sugi ./cmd/sugi
./sugi
```

Requires **Go 1.23+**.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues and PRs welcome!

---

## License

[MIT](LICENSE)
