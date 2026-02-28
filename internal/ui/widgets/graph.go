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
