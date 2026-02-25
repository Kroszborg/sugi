package ai

import "strings"

// CommitMsgPrompt builds a prompt to generate a conventional commit message.
func CommitMsgPrompt(diff string) string {
	// Truncate diff to avoid token limits
	if len(diff) > 4000 {
		diff = diff[:4000] + "\n... (truncated)"
	}
	return `You are an expert developer writing git commit messages following Conventional Commits.

OUTPUT FORMAT (exactly two parts separated by a blank line):
1. Subject line — max 72 chars: type(scope): present-tense description
2. Body — 1-3 sentences explaining WHY this change was made (not what the code does)

Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build
Rules:
- Subject must be lowercase after the colon, no trailing period
- Body is REQUIRED — always write it, even for small changes
- Body explains motivation, context, or impact — not a repeat of the subject
- NO markdown, NO code fences, NO "Here is...", NO preamble
- Output ONLY the two-part commit message, nothing else

Example output:
feat(auth): add JWT refresh token rotation

Refresh tokens were single-use and expired after 1h, causing frequent
logouts. Rotation extends sessions automatically while limiting the
damage window if a token is stolen.

Git diff:
` + diff
}

// PRDescriptionPrompt builds a prompt to generate a PR description.
func PRDescriptionPrompt(title string, commits []string, diff string) string {
	commitList := strings.Join(commits, "\n")
	if len(diff) > 3000 {
		diff = diff[:3000] + "\n... (truncated)"
	}
	return `You are an expert developer writing pull request descriptions.
Generate a clear, helpful PR description in Markdown format.

PR Title: ` + title + `

Commits in this PR:
` + commitList + `

Sample diff:
` + diff + `

Format:
## Summary
- Brief bullet points of what changed and why

## Changes
- More detailed technical bullet points

## Test plan
- How to verify this works

Write ONLY the description body (no title). Use GitHub-flavored Markdown.`
}

// DiffSummaryPrompt builds a prompt to explain a diff in plain English.
func DiffSummaryPrompt(diff string) string {
	if len(diff) > 4000 {
		diff = diff[:4000] + "\n... (truncated)"
	}
	return `You are a senior developer explaining code changes to a teammate.
Summarize the following git diff in 2-4 plain English sentences.
Focus on WHAT changed and WHY it matters. Be concise, no bullet points, no markdown.

Git diff:
` + diff
}

// BranchNamePrompt builds a prompt to suggest a branch name from staged changes.
func BranchNamePrompt(diff string) string {
	if len(diff) > 2000 {
		diff = diff[:2000] + "\n... (truncated)"
	}
	return `You are a developer naming a git branch.
Based on the following staged diff, suggest a single branch name.

Rules:
- Use kebab-case (lowercase, hyphens)
- Prefix with type: feat/, fix/, docs/, refactor/, chore/
- Max 50 chars total
- Be specific and descriptive
- Output ONLY the branch name, nothing else

Staged diff:
` + diff
}
