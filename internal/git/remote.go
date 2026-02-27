package git

import "strings"

// RemoteEntry holds full details for a single git remote.
type RemoteEntry struct {
	Name     string
	FetchURL string
	PushURL  string
}

// RemoteURL returns the URL for the named remote (e.g. "origin").
func (c *Client) RemoteURL(name string) (string, error) {
	return c.run("remote", "get-url", name)
}

// OriginURL returns the URL of the "origin" remote, or "" if none.
func (c *Client) OriginURL() string {
	url, err := c.RemoteURL("origin")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(url)
}

// Remotes returns a list of configured remotes.
func (c *Client) Remotes() ([]string, error) {
	return c.runLines("remote")
}

// ListRemotes returns all configured remotes with their URLs.
func (c *Client) ListRemotes() ([]RemoteEntry, error) {
	out, err := c.run("remote", "-v")
	if err != nil {
		return nil, err
	}
	return parseRemoteEntries(out), nil
}

// AddRemote adds a new named remote.
func (c *Client) AddRemote(name, url string) error {
	_, err := c.run("remote", "add", name, url)
	return err
}

// RemoveRemote removes a remote by name.
func (c *Client) RemoveRemote(name string) error {
	_, err := c.run("remote", "remove", name)
	return err
}

// RenameRemote renames a remote.
func (c *Client) RenameRemote(oldName, newName string) error {
	_, err := c.run("remote", "rename", oldName, newName)
	return err
}

// FetchRemote fetches a specific remote by name.
func (c *Client) FetchRemote(name string) error {
	_, err := c.run("fetch", name)
	return err
}

// parseRemoteEntries parses `git remote -v` output into RemoteEntry structs.
func parseRemoteEntries(raw string) []RemoteEntry {
	byName := make(map[string]*RemoteEntry)
	var order []string

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		tab := strings.Index(line, "\t")
		if tab < 0 {
			continue
		}
		name := line[:tab]
		rest := strings.TrimSpace(line[tab+1:])

		url, typ := rest, ""
		if idx := strings.LastIndex(rest, " ("); idx >= 0 {
			url = rest[:idx]
			typ = strings.Trim(rest[idx+2:], "()")
		}

		if _, exists := byName[name]; !exists {
			byName[name] = &RemoteEntry{Name: name}
			order = append(order, name)
		}
		r := byName[name]
		switch typ {
		case "fetch":
			r.FetchURL = url
		case "push":
			r.PushURL = url
		default:
			r.FetchURL = url
			r.PushURL = url
		}
	}

	out := make([]RemoteEntry, 0, len(order))
	for _, n := range order {
		out = append(out, *byName[n])
	}
	return out
}

// AheadBehind returns the number of commits ahead/behind the upstream for HEAD.
func (c *Client) AheadBehind() (ahead, behind int, err error) {
	out, err := c.run("rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(out)
	if len(parts) == 2 {
		ahead = atoi(parts[0])
		behind = atoi(parts[1])
	}
	return
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n = n*10 + int(r-'0')
		}
	}
	return n
}
