package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, []byte(templateConfig), 0o644)
		return cfg
	}
	if err != nil {
		return cfg
	}
	defer f.Close()

	// Partial decode: only override fields present in the file.
	// If groq_api_key is still the placeholder, treat it as empty.
	_ = json.NewDecoder(f).Decode(&cfg)
	if cfg.GroqAPIKey == "YOUR_GROQ_API_KEY_HERE" {
		cfg.GroqAPIKey = ""
	}
	return cfg
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
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
func readGHCLIToken(host string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Try ~/.config/gh/hosts.yml
	hostsPath := filepath.Join(homeDir, ".config", "gh", "hosts.yml")
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return ""
	}

	// Minimal YAML parsing - look for "oauth_token:" under the host
	content := string(data)
	hostIdx := indexOf(content, host)
	if hostIdx < 0 {
		return ""
	}
	tokenIdx := indexOf(content[hostIdx:], "oauth_token:")
	if tokenIdx < 0 {
		return ""
	}
	tokenLine := content[hostIdx+tokenIdx:]
	// Extract value after "oauth_token: "
	start := indexOf(tokenLine, "oauth_token:") + len("oauth_token:")
	if start >= len(tokenLine) {
		return ""
	}
	rest := tokenLine[start:]
	// Trim whitespace and take until newline
	i := 0
	for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t') {
		i++
	}
	rest = rest[i:]
	end := indexOf(rest, "\n")
	if end < 0 {
		end = len(rest)
	}
	token := rest[:end]
	// Strip quotes if present
	if len(token) >= 2 && token[0] == '"' && token[len(token)-1] == '"' {
		token = token[1 : len(token)-1]
	}
	return token
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
