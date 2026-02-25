package git

import "strings"

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
