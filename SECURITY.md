# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | ✅        |
| < 0.1   | ❌        |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

To report a security issue, use GitHub's private [security advisories](https://github.com/Kroszborg/sugi/security/advisories/new) (preferred) or open a private discussion.

Include as much detail as possible:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fix

You can expect a response within **48 hours** and a patch within **7 days** for confirmed critical issues.

## Scope

Items in scope for security reports:
- Arbitrary command execution from crafted git repo data
- Token / credential leakage (GitHub, GitLab, Groq API key)
- Path traversal or directory escape vulnerabilities
- Any vulnerability that allows a malicious git repository to compromise the user's system

Out of scope:
- Issues in upstream dependencies (report to the upstream maintainer)
- Theoretical vulnerabilities without a working proof of concept
- Social engineering

## Security Design Notes

sugi stores API keys in `~/.config/sugi/config.json` with `0o600` file permissions (owner read/write only). Keys are never logged or transmitted to any party other than the configured API endpoints (Groq, GitHub, GitLab).

sugi runs git commands using `os/exec` with argument lists — not shell interpolation — to avoid command injection. User-supplied strings (branch names, tag names, commit messages) are always passed as discrete arguments, never concatenated into shell strings.
