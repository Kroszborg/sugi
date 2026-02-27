package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all user-configurable settings for sugi.
// Stored at ~/.config/sugi/config.json
type Config struct {
	// AI — Groq (free at console.groq.com)
	GroqAPIKey string `json:"groq_api_key"` // or set GROQ_API_KEY env var
	GroqModel  string `json:"groq_model"`   // default: llama-3.1-8b-instant

	// Forge
	GitHubToken string `json:"github_token"` // overrides GITHUB_TOKEN env
	GitLabToken string `json:"gitlab_token"` // overrides GITLAB_TOKEN env
	GitLabHost  string `json:"gitlab_host"`  // for self-hosted GitLab

	// UI
	Theme        string `json:"theme"`         // dark (default)
	MouseEnabled bool   `json:"mouse_enabled"` // default: true
	ShowGraph    bool   `json:"show_graph"`    // default: false
}

// Default returns a config with sensible defaults.
func Default() Config {
	return Config{
		GroqModel:    "llama-3.1-8b-instant",
		Theme:        "dark",
		MouseEnabled: true,
		ShowGraph:    false,
	}
}

// templateConfig is written to disk on first run so users can see all options.
const templateConfig = `{
  "groq_api_key": "YOUR_GROQ_API_KEY_HERE",
  "groq_model": "llama-3.1-8b-instant",
  "github_token": "",
  "gitlab_token": "",
  "gitlab_host": "",
  "mouse_enabled": true,
  "show_graph": false
}
`

// Load reads the config file, returning defaults merged with any stored values.
// On first run it creates a template config so users know where to add their key.
func Load() Config {
	cfg := Default()

	path, err := configPath()
	if err != nil {
		return cfg
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		// First run: write a template config so the user can see where to put their key.
		// Use restrictive permissions — config contains API keys.
		if mkErr := os.MkdirAll(filepath.Dir(path), 0o700); mkErr == nil {
			_ = os.WriteFile(path, []byte(templateConfig), 0o600)
		}
		return cfg
	}
	if err != nil {
		return cfg
	}
	defer f.Close()

	// Partial decode: only override fields present in the file.
	// If groq_api_key is still the placeholder, treat it as empty.
	if decErr := json.NewDecoder(f).Decode(&cfg); decErr != nil {
		// Silently return defaults on JSON parse error — user may have partial config
		return Default()
	}
	if cfg.GroqAPIKey == "YOUR_GROQ_API_KEY_HERE" {
		cfg.GroqAPIKey = ""
	}
	return cfg
}

// Save writes the config to disk, creating the directory if needed.
// Uses restrictive file permissions (0o600) since config contains API keys.
func Save(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	// O_TRUNC to truncate existing file; 0o600 = owner read/write only
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

// ConfigDir returns the sugi config directory path.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "sugi"), nil
}

func configPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// GitHubToken returns the effective GitHub token: config > env.
func (c Config) EffectiveGitHubToken() string {
	if c.GitHubToken != "" {
		return c.GitHubToken
	}
	// Try environment
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	// Try GH_TOKEN (GitHub CLI style)
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return t
	}
	// Try reading from gh CLI config
	return readGHCLIToken("github.com")
}

// EffectiveGitLabToken returns the effective GitLab token.
func (c Config) EffectiveGitLabToken() string {
	if c.GitLabToken != "" {
		return c.GitLabToken
	}
	return os.Getenv("GITLAB_TOKEN")
}

// readGHCLIToken reads the token from the gh CLI config file.
// Parses hosts.yml line-by-line with proper host boundary detection to avoid
// matching substrings of other hostnames (e.g. "github" matching "mygithub.com").
func readGHCLIToken(host string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	hostsPath := filepath.Join(homeDir, ".config", "gh", "hosts.yml")
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return ""
	}

	// Line-by-line YAML parsing with strict host boundary detection.
	// gh CLI hosts.yml format:
	//   github.com:
	//       oauth_token: gho_xxx
	//       git_protocol: https
	inHost := false
	for _, line := range strings.Split(string(data), "\n") {
		// Skip comments and blank lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Top-level keys have no leading whitespace — host headers like "github.com:"
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			// Exact host match: key must be exactly "host:" with no extra chars
			key := strings.TrimSuffix(strings.TrimSpace(line), ":")
			inHost = (key == host)
			continue
		}

		// Inside the matched host block
		if inHost && strings.HasPrefix(trimmed, "oauth_token:") {
			token := strings.TrimSpace(strings.TrimPrefix(trimmed, "oauth_token:"))
			// Strip surrounding quotes if present
			if len(token) >= 2 && token[0] == '"' && token[len(token)-1] == '"' {
				token = token[1 : len(token)-1]
			}
			if token != "" {
				return token
			}
		}
	}
	return ""
}

