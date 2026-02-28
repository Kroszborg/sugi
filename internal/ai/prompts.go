package ai

import "strings"

// CommitMsgPrompt builds a prompt to generate a conventional commit message.
func CommitMsgPrompt(diff string) string {
	// Truncate diff to avoid token limits
	if len(diff) > 4000 {
		diff = diff[:4000] + "\n... (truncated)"
	}
	return `You are an expert developer writing git commit messages following Conventional Commits.
Analyze the git diff below and write a commit message for EXACTLY what changed in it.

OUTPUT FORMAT — two parts separated by one blank line:
1. Subject: type(scope): short description  (max 72 chars, lowercase after colon, no period)
2. Body: 1-3 sentences explaining WHY this change was made

Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build

STRICT RULES:
- Read the diff carefully — describe ONLY what is in the diff, nothing else
- Do NOT use any example from these instructions as your answer
- No markdown, no code fences, no preamble, no "Here is..."
- Output ONLY the commit message

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
