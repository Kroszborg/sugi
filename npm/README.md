# sugi

> Terminal UI git client — faster than lazygit, with AI commit messages and GitHub/GitLab PR integration.

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

Run inside any git repo:

```sh
sugi
```

## Keybindings

| Key | Action |
|-----|--------|
| `1` / `2` / `3` / `4` | Switch panels (Files / Branches / Commits / Diff) |
| `tab` | Next field |
| `s` | Stage / unstage file |
| `c` | Open commit form |
| `ctrl+g` / `alt+g` | Generate AI commit message (requires Groq API key) |
| `ctrl+s` | Commit |
| `b` | Blame view |
| `z` | Stash panel |
| `R` | Reflog |
| `O` | Settings (add Groq API key here) |
| `?` | Help |
| `q` | Quit |

## AI Commit Messages

sugi uses [Groq](https://console.groq.com) (free tier available) to generate conventional commit messages from your staged diff.

1. Press `O` to open Settings
2. Add your Groq API key
3. Stage your changes, open commit form with `c`
4. Press `ctrl+g` or `alt+g` to generate

## License

MIT © [Kroszborg](https://github.com/Kroszborg)
