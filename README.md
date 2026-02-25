# sugi 杉

> A terminal UI git client that beats lazygit — GitHub/GitLab PRs, AI commit messages, commit graph.

```
  sugi (杉) — cedar tree. Grows fast, stands tall, shaped with precision.
```

## Features

- **Files panel** — stage, unstage, discard, hunk-level staging (`space` / `u`)
- **Branches panel** — checkout, create, delete; upstream ahead/behind badges
- **Commits panel** — log, ASCII commit graph (`g`), blame (`b`), reflog (`R`)
- **Diff panel** — unified diff with hunk navigation (`[` / `]`), AI summary (`A`)
- **PR panel** — GitHub / GitLab pull requests with CI badges and review status (`P`)
- **Stash panel** — list, apply, pop, drop (`z`)
- **Commit form** — subject + body, 72-char counter, AI generation (`ctrl+g`)
- **Command palette** — fuzzy search over all actions (`ctrl+p`)
- **AI integration** — Ollama (local, private) → Groq (free cloud fallback)

## Install

```sh
# npm (any platform)
npm install -g sugi

# Homebrew (macOS / Linux)
brew install Kroszborg/tap/sugi

# go install
go install github.com/Kroszborg/sugi/cmd/sugi@latest
```

## Usage

```sh
sugi            # open in current directory
sugi /path/repo # open a specific repo
sugi version    # print version
```

## Key bindings

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | cycle panels |
| `1` `2` `3` `4` | jump to Files / Branches / Commits / Diff |
| `↑↓` / `j` `k` | navigate |
| `space` | stage / unstage file |
| `a` | stage all |
| `u` | unstage hunk (diff panel) |
| `c` | open commit form |
| `ctrl+g` | AI-generate commit message |
| `ctrl+s` | commit (in commit form) |
| `[` / `]` | prev / next hunk |
| `A` | AI summarise current diff |
| `P` | open pull requests |
| `z` | open stash |
| `b` | blame current file |
| `R` | reflog |
| `g` | toggle commit graph |
| `r` | refresh |
| `/` | search / filter commits |
| `ctrl+p` | command palette |
| `?` | help overlay |
| `q` | quit |

## AI setup (Groq — free)

1. Sign up at **[console.groq.com](https://console.groq.com)** (free, no credit card)
2. Create an API key
3. Add it to `~/.config/sugi/config.json`:

```json
{
  "groq_api_key": "gsk_..."
}
```

Or set the environment variable instead:
```sh
export GROQ_API_KEY=gsk_...
```

That's it. sugi uses `llama-3.1-8b-instant` by default — fast and free.
To use a different model add `"groq_model": "llama-3.3-70b-versatile"` to config.

## GitHub / GitLab

sugi auto-detects the forge from the `origin` remote URL.

**GitHub:** set `GITHUB_TOKEN` env var, or sugi reads from the `gh` CLI automatically.

**GitLab:** set `GITLAB_TOKEN` env var.

Or configure in `~/.config/sugi/config.json`:
```json
{
  "github_token": "ghp_...",
  "gitlab_token": "glpat-...",
  "gitlab_host": "https://gitlab.company.com"
}
```

## Config

`~/.config/sugi/config.json` — all fields are optional:

```json
{
  "groq_api_key": "gsk_...",
  "groq_model": "llama-3.1-8b-instant",
  "github_token": "",
  "gitlab_token": "",
  "gitlab_host": "",
  "mouse_enabled": true,
  "show_graph": false
}
```

## Build from source

```sh
git clone https://github.com/Kroszborg/sugi
cd sugi
go build -o sugi ./cmd/sugi
./sugi
```

Requires Go 1.23+.

## License

MIT
