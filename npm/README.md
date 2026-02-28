# sugi 杉

> A terminal UI git client — GitHub/GitLab PRs, AI commit messages, interactive rebase, bisect, worktrees, multi-account management, and more.

![sugi screenshot](https://github.com/Kroszborg/sugi/raw/master/public/image.png)

## Install

```sh
npm install -g sugi
```

Or via Homebrew:
```sh
brew install Kroszborg/tap/sugi
```

Or via Go:
```sh
go install github.com/Kroszborg/sugi/cmd/sugi@latest
```

## Usage

```sh
sugi              # open in current git repo
sugi /path/repo   # open a specific repo
sugi version      # print version
```

## Panels

| Panel | Key | Description |
|-------|-----|-------------|
| Files | `1` | Stage, unstage, discard, hunk-level staging, multi-select (`ctrl+space`) |
| Branches | `2` | Checkout, create, rename, delete, merge, rebase, open in browser |
| Commits | `3` | Log, graph, cherry-pick, revert, reset, interactive rebase, blame |
| Diff | `4` | Unified diff, hunk navigation, stage/unstage hunks, AI summary |
| Accounts | `A` | Manage multiple GitHub/GitLab named accounts |
| PRs / MRs | `P` | GitHub & GitLab pull requests with CI and review badges |
| Stash | `z` | List, apply, pop, drop stashes with diff preview |
| Blame | `b` | File blame per line — author, date, hash |
| Reflog | `R` | Reflog with undo capability |
| Worktrees | `W` | List, add, remove git worktrees |
| Remotes | `E` | List, add, remove, fetch remotes |
| Bisect | `B` | Interactive git bisect session |
| Interactive Rebase | `i` | Reorder/squash/fixup/drop commits |
| Conflict Resolver | auto | Opens on conflicted files — pick ours/theirs |
| File History | `L` | Commits touching the selected file |
| Command Palette | `ctrl+p`/`alt+p` | Fuzzy search all actions |
| Help | `?` | Scrollable keybinding reference |
| Settings | `O` | Edit config in-app, saved instantly |

## Key bindings

### Navigation
| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | cycle panels |
| `1` `2` `3` `4` | jump to panel |
| `j` `k` / `↑` `↓` | move up/down |
| `pgup` / `pgdn` | page up/down |
| `/` | search / filter |
| `esc` | back / cancel |

### Files
| Key | Action |
|-----|--------|
| `space` | stage / unstage |
| `ctrl+space` | multi-select |
| `a` | stage all |
| `d` | discard (with confirmation) |
| `c` | commit form |
| `ctrl+a` | amend HEAD |
| `P` / `p` | push / pull |
| `F` | force push with-lease |
| `Z` | stash all changes |
| `L` | file history |
| `s` | toggle staged/unstaged diff |

### Branches
| Key | Action |
|-----|--------|
| `enter` | checkout |
| `n` | new branch |
| `R` | rename branch |
| `D` | delete branch |
| `m` | merge into current |
| `r` | rebase onto branch |
| `o` | open on GitHub/GitLab |
| `P` / `p` | push / pull |

### Commits
| Key | Action |
|-----|--------|
| `y` | copy hash |
| `C` | cherry-pick |
| `v` | revert |
| `X` | reset HEAD (soft/mixed/hard) |
| `i` | interactive rebase |
| `o` | open on GitHub/GitLab |
| `g` | toggle graph |
| `b` | blame at this commit |
| `R` | reflog |
| `ctrl+a` | amend HEAD |

### Diff
| Key | Action |
|-----|--------|
| `[` / `]` | prev / next hunk |
| `space` | stage hunk |
| `u` | unstage hunk |
| `s` | toggle staged/unstaged |
| `ctrl+i` | AI-summarise diff |

### Accounts (`A` — from any panel)
| Key | Action |
|-----|--------|
| `tab` | GitHub / GitLab tab |
| `enter` | activate account |
| `n` | add new account |
| `D` | delete account |
| `esc` | close |

### Global
| Key | Action |
|-----|--------|
| `c` | commit form |
| `ctrl+g` / `alt+g` | AI-generate commit message |
| `ctrl+p` / `alt+p` | command palette |
| `O` | settings |
| `?` | help |
| `q` | quit |

## AI Commit Messages (free)

sugi uses [Groq](https://console.groq.com) to generate conventional commit messages from your staged diff.

1. Sign up at [console.groq.com](https://console.groq.com) (free, no credit card)
2. Press `O` in sugi → paste your API key in Settings
3. Stage your changes, open commit form with `c`
4. Press `ctrl+g` (or `alt+g` in VS Code terminal) to generate

## Multi-Account Management

Switch between personal and work GitHub/GitLab accounts without editing config:

1. Press `A` to open the Accounts panel
2. Press `n` to add a named account (name → token → optional custom host)
3. Press `enter` to activate — the active account appears in the status bar as `⬡ name`
4. Accounts persist across sessions in `~/.config/sugi/config.json`

## GitHub / GitLab

sugi auto-detects the forge from your `origin` remote.

**Simple (single account):**
- **GitHub:** set `GITHUB_TOKEN` or sugi reads from `gh` CLI automatically
- **GitLab:** set `GITLAB_TOKEN`

**Multiple accounts:** use the Accounts panel (`A`) described above.

## Config file

`~/.config/sugi/config.json` (auto-created on first run):

```json
{
  "groq_api_key": "gsk_...",
  "groq_model": "llama-3.1-8b-instant",
  "github_token": "ghp_...",
  "gitlab_token": "",
  "gitlab_host": "",
  "github_accounts": [
    { "name": "personal", "token": "ghp_..." },
    { "name": "work",     "token": "ghp_...", "host": "github.company.com" }
  ],
  "active_github_account": "work"
}
```

## License

MIT © [Kroszborg](https://github.com/Kroszborg)
