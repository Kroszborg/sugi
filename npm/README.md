# sugi 杉

> A terminal UI git client — GitHub/GitLab PRs, AI commit messages, interactive rebase, bisect, worktrees, and more.

![sugi screenshot](https://github.com/Kroszborg/sugi/raw/master/image.png)

## Install

```sh
npm install -g @kroszborg/sugi
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
sugi --version    # print version
```

## Panels

| Panel | Key | Description |
|-------|-----|-------------|
| Files | `1` | Stage, unstage, discard, hunk-level staging, multi-select |
| Branches | `2` | Checkout, create, rename, delete, merge, rebase |
| Commits | `3` | Log, graph, cherry-pick, revert, reset, interactive rebase |
| Diff | `4` | Unified diff, hunk navigation, side-by-side toggle |
| PRs / MRs | `p` | GitHub & GitLab pull requests with CI badges |
| Stash | `z` | List, apply, pop, drop stashes |
| Blame | `b` | File blame per line |
| Worktrees | `W` | List, add, remove git worktrees |
| Remotes | `E` | List, add, remove, fetch remotes |
| Bisect | `B` | Interactive git bisect session |
| Interactive Rebase | `i` | Reorder/squash/fixup/drop commits |
| File History | `L` | Commits touching the selected file |
| Command Palette | `ctrl+p` | Fuzzy search all actions |
| Help | `?` | Scrollable keybinding reference |
| Settings | `O` | Edit config in-app |

## Key bindings

### Navigation
| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | cycle panels |
| `1` `2` `3` `4` | jump to panel |
| `j` `k` / `↑` `↓` | move up/down |
| `g` / `G` | top / bottom |
| `esc` | back / cancel |

### Files
| Key | Action |
|-----|--------|
| `space` | stage / unstage |
| `a` | stage all |
| `u` | unstage |
| `d` | discard |
| `L` | file history |

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

### Commits
| Key | Action |
|-----|--------|
| `C` | cherry-pick |
| `v` | revert |
| `X` | reset HEAD (soft/mixed/hard) |
| `i` | interactive rebase |
| `o` | open on GitHub/GitLab |
| `g` | toggle graph |

### Global
| Key | Action |
|-----|--------|
| `c` | commit form |
| `ctrl+g` | AI-generate commit message |
| `P` | push |
| `p` | pull |
| `f` | fetch |
| `ctrl+p` | command palette |
| `?` | help |
| `q` | quit |

## AI Commit Messages (free)

sugi uses [Groq](https://console.groq.com) to generate conventional commit messages from your staged diff.

1. Sign up at [console.groq.com](https://console.groq.com) (free, no credit card)
2. Press `O` in sugi to open Settings and paste your API key
3. Stage your changes, open commit form with `c`
4. Press `ctrl+g` to generate

## GitHub / GitLab

sugi auto-detects the forge from your `origin` remote.

- **GitHub:** set `GITHUB_TOKEN` or sugi reads from `gh` CLI automatically
- **GitLab:** set `GITLAB_TOKEN`

## License

MIT © [Kroszborg](https://github.com/Kroszborg)
