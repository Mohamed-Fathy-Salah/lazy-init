package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var titleStyle = lipgloss.NewStyle().Bold(true)

func renderPanel(content string, width, height int, title string, focused bool) string {
	borderColor := lipgloss.Color("8")
	if focused {
		borderColor = lipgloss.Color("4")
	}

	bc := lipgloss.NewStyle().Foreground(borderColor)
	b := lipgloss.RoundedBorder()

	// Build top line with title
	titleStr := ""
	if title != "" {
		titleStr = titleStyle.Foreground(borderColor).Render(" " + title + " ")
	}
	innerWidth := width - 2
	topLine := bc.Render(b.TopLeft) + titleStr + bc.Render(strings.Repeat(b.Top, max(0, innerWidth-lipgloss.Width(titleStr)))) + bc.Render(b.TopRight)

	// Build bottom line
	bottomLine := bc.Render(b.BottomLeft) + bc.Render(strings.Repeat(b.Bottom, innerWidth)) + bc.Render(b.BottomRight)

	// Render content with padding
	contentStyle := lipgloss.NewStyle().Width(innerWidth).Height(height - 2)
	body := contentStyle.Render(content)

	// Add side borders to each line
	var lines []string
	lines = append(lines, topLine)
	for _, line := range strings.Split(body, "\n") {
		lines = append(lines, bc.Render(b.Left)+line+bc.Render(b.Right))
	}
	lines = append(lines, bottomLine)

	return strings.Join(lines, "\n")
}
