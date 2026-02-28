package forge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GitLabClient implements ForgeClient for GitLab.com and self-hosted GitLab.
type GitLabClient struct {
	info   ForgeInfo
	token  string
	apiURL string
	http   *http.Client
}

// NewGitLabClient creates a GitLab API client.
func NewGitLabClient(info ForgeInfo, token string) *GitLabClient {
	return &GitLabClient{
		info:   info,
		token:  token,
		apiURL: info.APIURL(),
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *GitLabClient) ForgeInfo() ForgeInfo { return c.info }

// projectPath returns URL-encoded "owner/repo".
func (c *GitLabClient) projectPath() string {
	// GitLab uses URL encoding for namespace/project
	return fmt.Sprintf("%s%%2F%s", c.info.Owner, c.info.Repo)
}

// --- GitLab API response types ---

type glMR struct {
	IID         int     `json:"iid"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	State       string  `json:"state"`
	Draft       bool    `json:"draft"`
	WebURL      string  `json:"web_url"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	MergedAt    *string `json:"merged_at"`
	Author      struct {
		Username string `json:"username"`
	} `json:"author"`
	SourceBranch string   `json:"source_branch"`
	TargetBranch string   `json:"target_branch"`
	Labels       []string `json:"labels"`
}

type glPipeline struct {
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

// --- ForgeClient implementation ---

func (c *GitLabClient) ListPRs(state string) ([]PullRequest, error) {
	glState := "opened"
	if state == "closed" {
		glState = "closed"
	} else if state == "merged" {
		glState = "merged"
	}

	url := fmt.Sprintf("%s/projects/%s/merge_requests?state=%s&per_page=50",
		c.apiURL, c.projectPath(), glState)

	var mrs []glMR
	if err := c.get(url, &mrs); err != nil {
		return nil, err
	}

	prs := make([]PullRequest, len(mrs))
	for i, mr := range mrs {
		prs[i] = c.convertMR(mr)
	}
	return prs, nil
}

func (c *GitLabClient) GetPR(number int) (*PullRequest, error) {
	url := fmt.Sprintf("%s/projects/%s/merge_requests/%d",
		c.apiURL, c.projectPath(), number)
	var mr glMR
	if err := c.get(url, &mr); err != nil {
		return nil, err
	}
	pr := c.convertMR(mr)
	return &pr, nil
}

func (c *GitLabClient) GetPRForBranch(branch string) (*PullRequest, error) {
	url := fmt.Sprintf("%s/projects/%s/merge_requests?source_branch=%s&state=opened",
		c.apiURL, c.projectPath(), branch)
	var mrs []glMR
	if err := c.get(url, &mrs); err != nil {
		return nil, err
	}
	if len(mrs) == 0 {
		return nil, nil
	}
	pr := c.convertMR(mrs[0])
	return &pr, nil
}

func (c *GitLabClient) CreatePR(title, body, head, base string, draft bool) (*PullRequest, error) {
	titleStr := title
	if draft {
		titleStr = "Draft: " + title
	}
	payload := map[string]interface{}{
		"title":         titleStr,
		"description":   body,
		"source_branch": head,
		"target_branch": base,
	}
	url := fmt.Sprintf("%s/projects/%s/merge_requests", c.apiURL, c.projectPath())
	var mr glMR
	if err := c.post(url, payload, &mr); err != nil {
		return nil, err
	}
	pr := c.convertMR(mr)
	return &pr, nil
}

func (c *GitLabClient) MergePR(number int, _ string) error {
	url := fmt.Sprintf("%s/projects/%s/merge_requests/%d/merge",
		c.apiURL, c.projectPath(), number)
	return c.put(url, map[string]string{})
}

func (c *GitLabClient) ClosePR(number int) error {
	url := fmt.Sprintf("%s/projects/%s/merge_requests/%d",
		c.apiURL, c.projectPath(), number)
	return c.doJSON("PUT", url, map[string]string{"state_event": "close"}, nil)
}

func (c *GitLabClient) GetCIStatus(commitSHA string) (*CIStatus, error) {
	url := fmt.Sprintf("%s/projects/%s/repository/commits/%s/statuses",
		c.apiURL, c.projectPath(), commitSHA)

	var statuses []struct {
		Name      string `json:"name"`
		Status    string `json:"status"`
		TargetURL string `json:"target_url"`
	}
	if err := c.get(url, &statuses); err != nil {
		return &CIStatus{Result: CINone}, nil
	}

	ci := &CIStatus{}
	for _, s := range statuses {
		ci.Checks = append(ci.Checks, CICheck{
			Name:       s.Name,
			Status:     s.Status,
			Conclusion: s.Status,
			URL:        s.TargetURL,
		})
	}
	ci.Result = aggregateCIResult(ci.Checks)
	return ci, nil
}

// --- Helpers ---

func (c *GitLabClient) convertMR(mr glMR) PullRequest {
	state := PROpen
	switch mr.State {
	case "closed":
		state = PRClosed
	case "merged":
		state = PRMerged
	}
	if mr.Draft {
		state = PRDraft
	}

	created, _ := time.Parse(time.RFC3339, mr.CreatedAt)
	updated, _ := time.Parse(time.RFC3339, mr.UpdatedAt)

	var mergedAt *time.Time
	if mr.MergedAt != nil && *mr.MergedAt != "" {
		t, _ := time.Parse(time.RFC3339, *mr.MergedAt)
		mergedAt = &t
	}

	return PullRequest{
		Number:     mr.IID,
		Title:      mr.Title,
		Body:       mr.Description,
		State:      state,
		Author:     mr.Author.Username,
		HeadBranch: mr.SourceBranch,
		BaseBranch: mr.TargetBranch,
		URL:        mr.WebURL,
		CreatedAt:  created,
		UpdatedAt:  updated,
		MergedAt:   mergedAt,
		Labels:     mr.Labels,
		IsDraft:    mr.Draft,
	}
}

// --- HTTP helpers ---

func (c *GitLabClient) get(url string, out interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("GitLab API %s: %s", resp.Status, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *GitLabClient) post(url string, body, out interface{}) error {
	return c.doJSON("POST", url, body, out)
}

func (c *GitLabClient) put(url string, body interface{}) error {
	return c.doJSON("PUT", url, body, nil)
}

func (c *GitLabClient) doJSON(method, url string, body, out interface{}) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("GitLab API %s: %s", resp.Status, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *GitLabClient) setHeaders(req *http.Request) {
	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}
	req.Header.Set("User-Agent", "sugi/1.0")
}
