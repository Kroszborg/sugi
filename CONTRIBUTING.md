# Contributing to sugi

Thank you for your interest in contributing! sugi is a terminal UI git client written in Go.

## Getting Started

### Prerequisites
- Go 1.24+
- A Unix-like terminal (Windows Terminal, iTerm2, etc.)

### Development setup

```bash
git clone https://github.com/Kroszborg/sugi
cd sugi
go build -o sugi.exe ./cmd/sugi   # Windows
go build -o sugi ./cmd/sugi       # macOS / Linux
./sugi                            # run in any git repo
```

### Project layout

```
cmd/sugi/           entry point (cobra CLI)
internal/
  git/              all git operations (exec.Command wrappers)
  ui/               bubbletea TUI
    panels/         files, branches, commits, diff, tags, blame, stash, reflog, PR panels
    widgets/        scrolllist, modal, statusbar, helpoverlay, badge
  ai/               Groq streaming client + prompts
  config/           JSON config at ~/.config/sugi/config.json
  forge/            GitHub / GitLab API clients
```

## Making Changes

1. **Fork** the repo and create a branch: `git checkout -b feat/my-feature`
2. **Write** your code. Run `go build ./...` and `go vet ./...` before committing.
3. **Test** manually — run `./sugi` in a repo with staged/unstaged changes.
4. **Commit** with a conventional commit message (sugi AI can generate it for you!):
   - `feat:` new functionality
   - `fix:` bug fix
   - `refactor:` code change without behaviour change
   - `docs:` documentation only
5. **Push** and open a pull request against `master`.

## Code Style

- Run `gofmt` before committing (or use `goimports`).
- Keep functions small — prefer composition over large switch blocks.
- New panels should follow the existing pattern in `internal/ui/panels/`.
- New git operations belong in `internal/git/` and should use the `c.run()` helper.
- Avoid external dependencies unless strictly necessary.

## Reporting Issues

Please use the [GitHub issue templates](.github/ISSUE_TEMPLATE/) when filing bugs or feature requests. Include:
- OS and terminal emulator
- Go version (`go version`)
- Steps to reproduce
- Expected vs actual behaviour

## Security

If you find a security vulnerability, **do not open a public issue**. See [SECURITY.md](SECURITY.md) for the responsible disclosure process.

## License

By contributing you agree that your contributions will be licensed under the [MIT License](LICENSE).
