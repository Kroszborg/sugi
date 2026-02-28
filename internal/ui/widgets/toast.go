package widgets

import (
	"github.com/charmbracelet/lipgloss"
)

// Toast holds a single notification message.
type Toast struct {
	Text  string
	IsErr bool
}

// ToastQueue manages a FIFO queue of toast notifications.
type ToastQueue struct {
	items []Toast
}

// Push adds a new toast to the queue.
func (q *ToastQueue) Push(text string, isErr bool) {
	q.items = append(q.items, Toast{Text: text, IsErr: isErr})
}

// Current returns the front-of-queue toast, or nil if empty.
func (q *ToastQueue) Current() *Toast {
	if len(q.items) == 0 {
		return nil
	}
	return &q.items[0]
}

// Pop removes the front toast.
func (q *ToastQueue) Pop() {
	if len(q.items) > 0 {
		q.items = q.items[1:]
	}
}

// IsEmpty returns true when no toasts are queued.
func (q *ToastQueue) IsEmpty() bool { return len(q.items) == 0 }

// View renders the current toast as a styled string.
// Returns "" when the queue is empty.
func (q *ToastQueue) View(width int) string {
	t := q.Current()
	if t == nil {
		return ""
	}
	color := "#3ecf8e"
	if t.IsErr {
		color = "#e05454"
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a2a")).
		Foreground(lipgloss.Color(color)).
		Padding(0, 2).
		Width(width).
		Render("  " + t.Text)
}
