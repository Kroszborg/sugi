package forge

import (
	"strings"
)

// ForgeType identifies the hosting platform.
type ForgeType int

const (
	ForgeUnknown ForgeType = iota
	ForgeGitHub
	ForgeGitLab
	ForgeGitea
)

func (f ForgeType) String() string {
	switch f {
	case ForgeGitHub:
		return "GitHub"
	case ForgeGitLab:
		return "GitLab"
	case ForgeGitea:
		return "Gitea"
	default:
		return "Unknown"
	}
}

// ForgeInfo holds the detected forge type and repo coordinates.
type ForgeInfo struct {
	Type  ForgeType
	Host  string // e.g. "github.com" or "gitlab.company.com"
	Owner string // e.g. "Kroszborg"
	Repo  string // e.g. "sugi"
}

// IsKnown returns true if the forge was detected.
func (f ForgeInfo) IsKnown() bool {
	return f.Type != ForgeUnknown && f.Owner != "" && f.Repo != ""
}

// APIURL returns the API base URL for this forge.
func (f ForgeInfo) APIURL() string {
	switch f.Type {
	case ForgeGitHub:
		if f.Host == "github.com" {
			return "https://api.github.com"
		}
		// GitHub Enterprise
		return "https://" + f.Host + "/api/v3"
	case ForgeGitLab:
		if f.Host == "gitlab.com" {
			return "https://gitlab.com/api/v4"
		}
		return "https://" + f.Host + "/api/v4"
	default:
		return ""
	}
}

// Detect parses a git remote URL and returns forge information.
// Supports SSH (git@github.com:owner/repo.git) and HTTPS (https://github.com/owner/repo).
func Detect(remoteURL string) ForgeInfo {
	remoteURL = strings.TrimSpace(remoteURL)
	if remoteURL == "" {
		return ForgeInfo{}
	}

	var host, path string

	if strings.HasPrefix(remoteURL, "git@") {
		// SSH: git@github.com:owner/repo.git
		rest := strings.TrimPrefix(remoteURL, "git@")
		idx := strings.Index(rest, ":")
		if idx < 0 {
			return ForgeInfo{}
		}
		host = rest[:idx]
		path = rest[idx+1:]
	} else if strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://") {
		// HTTPS: https://github.com/owner/repo.git
		withoutScheme := remoteURL
		if strings.HasPrefix(withoutScheme, "https://") {
			withoutScheme = strings.TrimPrefix(withoutScheme, "https://")
		} else {
			withoutScheme = strings.TrimPrefix(withoutScheme, "http://")
		}
		// Strip optional user:pass@
		if atIdx := strings.Index(withoutScheme, "@"); atIdx >= 0 {
			withoutScheme = withoutScheme[atIdx+1:]
		}
		slashIdx := strings.Index(withoutScheme, "/")
		if slashIdx < 0 {
			return ForgeInfo{}
		}
		host = withoutScheme[:slashIdx]
		path = withoutScheme[slashIdx+1:]
	} else {
		return ForgeInfo{}
	}

	// Strip .git suffix
	path = strings.TrimSuffix(path, ".git")
	path = strings.Trim(path, "/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return ForgeInfo{}
	}
	owner, repo := parts[0], parts[1]

	forgeType := classifyHost(host)
	return ForgeInfo{
		Type:  forgeType,
		Host:  host,
		Owner: owner,
		Repo:  repo,
	}
}

func classifyHost(host string) ForgeType {
	h := strings.ToLower(host)
	switch {
	case h == "github.com" || strings.HasSuffix(h, ".github.com") || strings.HasSuffix(h, ".ghe.com"):
		return ForgeGitHub
	case h == "gitlab.com" || strings.Contains(h, "gitlab"):
		return ForgeGitLab
	case strings.Contains(h, "gitea") || strings.Contains(h, "codeberg"):
		return ForgeGitea
	default:
		return ForgeUnknown
	}
}
