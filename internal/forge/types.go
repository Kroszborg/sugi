package forge

import "time"

// PRState is the current state of a pull request.
type PRState string

const (
	PROpen   PRState = "open"
	PRClosed PRState = "closed"
	PRMerged PRState = "merged"
	PRDraft  PRState = "draft"
)

// CIResult is the overall result of CI checks.
type CIResult string

const (
	CIPending CIResult = "pending"
	CISuccess CIResult = "success"
	CIFailure CIResult = "failure"
	CIError   CIResult = "error"
	CINone    CIResult = ""
)

// ReviewState is the state of a code review.
type ReviewState string

const (
	ReviewPending          ReviewState = "pending"
	ReviewApproved         ReviewState = "approved"
	ReviewChangesRequested ReviewState = "changes_requested"
	ReviewCommented        ReviewState = "commented"
	ReviewDismissed        ReviewState = "dismissed"
)

// PullRequest is a forge-agnostic PR/MR.
type PullRequest struct {
	Number     int
	Title      string
	Body       string
	State      PRState
	Author     string
	HeadBranch string
	BaseBranch string
	URL        string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	MergedAt   *time.Time
	CI         CIResult
	Reviews    []Review
	Labels     []string
	IsDraft    bool
}

// Review is a single code review.
type Review struct {
	Author    string
	State     ReviewState
	Body      string
	CreatedAt time.Time
}

// CICheck is a single CI check run.
type CICheck struct {
	Name       string
	Status     string // "queued", "in_progress", "completed"
	Conclusion string // "success", "failure", "cancelled", etc.
	URL        string
}

// CIStatus aggregates CI checks for a commit or PR.
type CIStatus struct {
	Result CIResult
	Checks []CICheck
}

// ForgeClient is the interface all forge backends implement.
type ForgeClient interface {
	// PRs
	ListPRs(state string) ([]PullRequest, error)
	GetPR(number int) (*PullRequest, error)
	GetPRForBranch(branch string) (*PullRequest, error)
	CreatePR(title, body, head, base string, draft bool) (*PullRequest, error)
	MergePR(number int, method string) error
	ClosePR(number int) error

	// CI
	GetCIStatus(commitSHA string) (*CIStatus, error)

	// Info
	ForgeInfo() ForgeInfo
}
