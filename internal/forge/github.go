package forge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GitHubClient implements ForgeClient for GitHub and GitHub Enterprise.
type GitHubClient struct {
	info   ForgeInfo
	token  string
	apiURL string
	http   *http.Client
}

// NewGitHubClient creates a GitHub API client.
func NewGitHubClient(info ForgeInfo, token string) *GitHubClient {
	return &GitHubClient{
		info:   info,
		token:  token,
		apiURL: info.APIURL(),
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *GitHubClient) ForgeInfo() ForgeInfo { return c.info }

// --- GitHub API response types ---

type ghPR struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	Draft     bool   `json:"draft"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	MergedAt  *string `json:"merged_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

type ghCheckRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
}

type ghCommitStatus struct {
	State    string `json:"state"`
	Statuses []struct {
		State       string `json:"state"`
		Context     string `json:"context"`
		Description string `json:"description"`
		TargetURL   string `json:"target_url"`
	} `json:"statuses"`
}

type ghReview struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	State       string `json:"state"`
	Body        string `json:"body"`
	SubmittedAt string `json:"submitted_at"`
}

// --- ForgeClient implementation ---

func (c *GitHubClient) ListPRs(state string) ([]PullRequest, error) {
	if state == "" {
		state = "open"
	}
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s&per_page=50",
		c.apiURL, c.info.Owner, c.info.Repo, state)

	var ghPRs []ghPR
	if err := c.get(url, &ghPRs); err != nil {
		return nil, err
	}

	prs := make([]PullRequest, len(ghPRs))
	for i, gp := range ghPRs {
		prs[i] = c.convertPR(gp)
	}
	return prs, nil
}

func (c *GitHubClient) GetPR(number int) (*PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d",
		c.apiURL, c.info.Owner, c.info.Repo, number)
	var gp ghPR
	if err := c.get(url, &gp); err != nil {
		return nil, err
	}
	pr := c.convertPR(gp)
	// Also fetch reviews
	reviews, _ := c.getReviews(number)
	pr.Reviews = reviews
	return &pr, nil
}

func (c *GitHubClient) GetPRForBranch(branch string) (*PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?head=%s:%s&state=open",
		c.apiURL, c.info.Owner, c.info.Repo, c.info.Owner, branch)
	var ghPRs []ghPR
	if err := c.get(url, &ghPRs); err != nil {
		return nil, err
	}
	if len(ghPRs) == 0 {
		return nil, nil
	}
	pr := c.convertPR(ghPRs[0])
	return &pr, nil
}

func (c *GitHubClient) CreatePR(title, body, head, base string, draft bool) (*PullRequest, error) {
	payload := map[string]interface{}{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
		"draft": draft,
	}
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", c.apiURL, c.info.Owner, c.info.Repo)
	var gp ghPR
	if err := c.post(url, payload, &gp); err != nil {
		return nil, err
	}
	pr := c.convertPR(gp)
	return &pr, nil
}

func (c *GitHubClient) MergePR(number int, method string) error {
	if method == "" {
		method = "merge"
	}
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/merge", c.apiURL, c.info.Owner, c.info.Repo, number)
	return c.put(url, map[string]string{"merge_method": method})
}

func (c *GitHubClient) ClosePR(number int) error {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.apiURL, c.info.Owner, c.info.Repo, number)
	return c.patch(url, map[string]string{"state": "closed"})
}

func (c *GitHubClient) GetCIStatus(commitSHA string) (*CIStatus, error) {
	// Fetch check runs (GitHub Actions)
	checkURL := fmt.Sprintf("%s/repos/%s/%s/commits/%s/check-runs",
		c.apiURL, c.info.Owner, c.info.Repo, commitSHA)

	var checkResp struct {
		CheckRuns []ghCheckRun `json:"check_runs"`
	}
	_ = c.get(checkURL, &checkResp)

	// Also fetch legacy commit statuses
	statusURL := fmt.Sprintf("%s/repos/%s/%s/commits/%s/status",
		c.apiURL, c.info.Owner, c.info.Repo, commitSHA)
	var legacyStatus ghCommitStatus
	_ = c.get(statusURL, &legacyStatus)

	return buildCIStatus(checkResp.CheckRuns, legacyStatus), nil
}

// --- Helpers ---

func (c *GitHubClient) getReviews(prNumber int) ([]Review, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews",
		c.apiURL, c.info.Owner, c.info.Repo, prNumber)
	var ghReviews []ghReview
	if err := c.get(url, &ghReviews); err != nil {
		return nil, err
	}
	reviews := make([]Review, len(ghReviews))
	for i, r := range ghReviews {
		t, _ := time.Parse(time.RFC3339, r.SubmittedAt)
		reviews[i] = Review{
			Author:    r.User.Login,
			State:     ReviewState(r.State),
			Body:      r.Body,
			CreatedAt: t,
		}
	}
	return reviews, nil
}

func (c *GitHubClient) convertPR(gp ghPR) PullRequest {
	state := PRState(gp.State)
	if gp.Draft {
		state = PRDraft
	}
	if gp.MergedAt != nil && *gp.MergedAt != "" {
		state = PRMerged
	}

	created, _ := time.Parse(time.RFC3339, gp.CreatedAt)
	updated, _ := time.Parse(time.RFC3339, gp.UpdatedAt)

	var mergedAt *time.Time
	if gp.MergedAt != nil && *gp.MergedAt != "" {
		t, _ := time.Parse(time.RFC3339, *gp.MergedAt)
		mergedAt = &t
	}

	labels := make([]string, len(gp.Labels))
	for i, l := range gp.Labels {
		labels[i] = l.Name
	}

	return PullRequest{
		Number:     gp.Number,
		Title:      gp.Title,
		Body:       gp.Body,
		State:      state,
		Author:     gp.User.Login,
		HeadBranch: gp.Head.Ref,
		BaseBranch: gp.Base.Ref,
		URL:        gp.HTMLURL,
		CreatedAt:  created,
		UpdatedAt:  updated,
		MergedAt:   mergedAt,
		Labels:     labels,
		IsDraft:    gp.Draft,
	}
}

func buildCIStatus(checks []ghCheckRun, legacy ghCommitStatus) *CIStatus {
	ci := &CIStatus{}

	// Process check runs
	for _, cr := range checks {
		check := CICheck{
			Name:       cr.Name,
			Status:     cr.Status,
			Conclusion: cr.Conclusion,
			URL:        cr.HTMLURL,
		}
		ci.Checks = append(ci.Checks, check)
	}

	// Process legacy statuses
	for _, s := range legacy.Statuses {
		ci.Checks = append(ci.Checks, CICheck{
			Name:       s.Context,
			Status:     "completed",
			Conclusion: s.State,
			URL:        s.TargetURL,
		})
	}

	// Aggregate result
	ci.Result = aggregateCIResult(ci.Checks)
	return ci
}

func aggregateCIResult(checks []CICheck) CIResult {
	if len(checks) == 0 {
		return CINone
	}
	hasFailure := false
	hasPending := false
	for _, c := range checks {
		switch {
		case c.Status == "in_progress" || c.Status == "queued":
			hasPending = true
		case c.Conclusion == "failure" || c.Conclusion == "error" || c.Conclusion == "timed_out":
			hasFailure = true
		}
	}
	if hasFailure {
		return CIFailure
	}
	if hasPending {
		return CIPending
	}
	return CISuccess
}

// --- HTTP helpers ---

func (c *GitHubClient) get(url string, out interface{}) error {
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
		return fmt.Errorf("GitHub API %s: %s", resp.Status, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *GitHubClient) post(url string, body, out interface{}) error {
	return c.doJSON("POST", url, body, out)
}

func (c *GitHubClient) put(url string, body interface{}) error {
	return c.doJSON("PUT", url, body, nil)
}

func (c *GitHubClient) patch(url string, body interface{}) error {
	return c.doJSON("PATCH", url, body, nil)
}

func (c *GitHubClient) doJSON(method, url string, body, out interface{}) error {
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
		return fmt.Errorf("GitHub API %s: %s", resp.Status, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *GitHubClient) setHeaders(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "sugi/1.0")
}
