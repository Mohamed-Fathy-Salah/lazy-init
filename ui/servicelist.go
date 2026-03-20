package ui

import (
	"fmt"
	"lazy-init/core"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	statusDown    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	statusUnknown = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	cursorStyle   = lipgloss.NewStyle().Bold(true).Reverse(true)
)

type serviceListModel struct {
	services []core.Service
	cursor   int
	offset   int
	width    int
	height   int
}

func (m *serviceListModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *serviceListModel) Selected() string {
	if len(m.services) == 0 {
		return ""
	}
	return m.services[m.cursor].Name
}

func (m *serviceListModel) Up() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
	}
}

func (m *serviceListModel) Down() {
	if m.cursor < len(m.services)-1 {
		m.cursor++
		visible := m.height - 2 // border
		if m.cursor >= m.offset+visible {
			m.offset = m.cursor - visible + 1
		}
	}
}

func (m *serviceListModel) View(focused bool) string {
	visible := m.height - 2 // border
	if visible < 1 {
		return ""
	}

	var lines []string
	end := m.offset + visible
	if end > len(m.services) {
		end = len(m.services)
	}

	for i := m.offset; i < end; i++ {
		svc := m.services[i]
		dot := statusDot(svc.Status)
		name := fmt.Sprintf("%-*s %s", m.width-12, svc.Name, svc.Status)
		if i == m.cursor {
			name = cursorStyle.Render(name)
		}
		lines = append(lines, dot+" "+name)
	}

	// Pad remaining lines
	for len(lines) < visible {
		lines = append(lines, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	border := lipgloss.NormalBorder()
	style := lipgloss.NewStyle().
		Border(border).
		Width(m.width - 2).
		Height(visible).
		BorderForeground(lipgloss.Color("8"))

	if focused {
		style = style.BorderForeground(lipgloss.Color("4"))
	}

	return style.Render(content)
}

func statusDot(status core.Status) string {
	switch status {
	case "running", "active":
		return statusRunning.Render("●")
	case "down", "inactive":
		return statusDown.Render("●")
	case "failed":
		return statusDown.Bold(true).Render("●")
	default:
		return statusUnknown.Render("●")
	}
}
