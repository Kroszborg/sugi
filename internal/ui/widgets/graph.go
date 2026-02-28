package widgets

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// GraphRenderer colors git log --graph output lines.
type GraphRenderer struct {
	colors       []lipgloss.Color
	hashStyle    lipgloss.Style
	subjectStyle lipgloss.Style
}

// NewGraphRenderer creates a renderer for commit graph lines.
func NewGraphRenderer() GraphRenderer {
	return GraphRenderer{
		colors: []lipgloss.Color{
			lipgloss.Color("#4d9de0"), // blue
			lipgloss.Color("#3ecf8e"), // green
			lipgloss.Color("#a87efb"), // purple
			lipgloss.Color("#d4a017"), // yellow
			lipgloss.Color("#e8835c"), // orange
			lipgloss.Color("#2ec4b6"), // teal
			lipgloss.Color("#7c6dfa"), // sky
		},
		hashStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#e8835c")).Bold(true),
		subjectStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8ee")),
	}
}

// RenderLines colorizes a slice of git log --graph lines.
// Each line is like "* abc1234 subject" or "| * def5678 subject" or graph-only lines.
func (g GraphRenderer) RenderLines(lines []string, width int) []string {
	rendered := make([]string, len(lines))
	for i, line := range lines {
		rendered[i] = g.renderLine(line, width)
	}
	return rendered
}

// ExtractHashes returns the short commit hash for each graph line, or "" for
// graph-only lines (connectors, blanks). Index aligns 1:1 with the lines slice.
func (g GraphRenderer) ExtractHashes(lines []string) []string {
	hashes := make([]string, len(lines))
	for i, line := range lines {
		hashes[i] = extractHashFromGraphLine(line)
	}
	return hashes
}

// extractHashFromGraphLine pulls the first 7-char hex token after graph chars.
func extractHashFromGraphLine(line string) string {
	// Skip graph prefix (*, |, /, \, _, space)
	j := 0
	for j < len(line) && isGraphChar(line[j]) {
		j++
	}
	rest := strings.TrimSpace(line[j:])
	if rest == "" {
		return ""
	}
	// First token is the short hash
	parts := strings.SplitN(rest, " ", 2)
	h := parts[0]
	// Validate: must be hex chars, 6-40 chars
	if len(h) < 6 || len(h) > 40 {
		return ""
	}
	for _, c := range h {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return ""
		}
	}
	return h
}

// renderLine colorizes a single graph line.
func (g GraphRenderer) renderLine(line string, width int) string {
	if line == "" {
		return ""
	}

	// Separate graph prefix from commit info
	// Graph chars: * | / \ _ space
	graphEnd := 0
	for graphEnd < len(line) {
		ch := line[graphEnd]
		if isGraphChar(ch) {
			graphEnd++
		} else {
			break
		}
	}

	graphPart := line[:graphEnd]
	rest := ""
	if graphEnd < len(line) {
		rest = line[graphEnd:]
	}

	// Color the graph characters
	coloredGraph := g.colorizeGraph(graphPart)

	if rest == "" {
		return coloredGraph
	}

	// Parse "hash subject" from rest
	parts := strings.SplitN(strings.TrimSpace(rest), " ", 2)
	if len(parts) == 0 {
		return coloredGraph + rest
	}

	hash := g.hashStyle.Render(parts[0])
	subject := ""
	if len(parts) > 1 {
		sub := parts[1]
		// Truncate subject to fit width
		maxSubject := width - graphEnd - 9 - 2
		if maxSubject > 5 && len(sub) > maxSubject {
			sub = sub[:maxSubject-1] + "…"
		}
		subject = " " + g.subjectStyle.Render(sub)
	}

	return coloredGraph + hash + subject
}

// colorizeGraph applies colors to graph prefix characters.
func (g GraphRenderer) colorizeGraph(graph string) string {
	if graph == "" {
		return ""
	}

	var sb strings.Builder
	colorIdx := 0

	for _, ch := range graph {
		switch ch {
		case '*':
			sb.WriteString(lipgloss.NewStyle().Foreground(g.colors[colorIdx%len(g.colors)]).Bold(true).Render("*"))
			colorIdx++
		case '|':
			sb.WriteString(lipgloss.NewStyle().Foreground(g.colors[colorIdx%len(g.colors)]).Render("|"))
			colorIdx++
		case '/', '\\':
			sb.WriteString(lipgloss.NewStyle().Foreground(g.colors[colorIdx%len(g.colors)]).Render(string(ch)))
			colorIdx++
		default:
			sb.WriteRune(ch)
		}
	}

	return sb.String()
}

func isGraphChar(ch byte) bool {
	return ch == '*' || ch == '|' || ch == '/' || ch == '\\' ||
		ch == '_' || ch == ' ' || ch == '-'
}
