package ui

import (
	"fmt"
	"lazy-init/core"

	"github.com/charmbracelet/lipgloss"
)

var (
	nameRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	nameDown    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	nameUnknown = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	nameFailed  = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cursorStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
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
	if len(m.services) == 0 {
		return
	}
	if m.cursor > 0 {
		m.cursor--
	} else {
		m.cursor = len(m.services) - 1
	}
	m.fixOffset()
}

func (m *serviceListModel) Down() {
	if len(m.services) == 0 {
		return
	}
	if m.cursor < len(m.services)-1 {
		m.cursor++
	} else {
		m.cursor = 0
	}
	m.fixOffset()
}

func (m *serviceListModel) GoTop() {
	m.cursor = 0
	m.offset = 0
}

func (m *serviceListModel) GoBottom() {
	if len(m.services) == 0 {
		return
	}
	m.cursor = len(m.services) - 1
	m.fixOffset()
}

func (m *serviceListModel) fixOffset() {
	visible := m.height - 2 // border
	if visible < 1 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
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
		line := fmt.Sprintf("%-*s", m.width-4, svc.Name)

		if i == m.cursor {
			line = cursorStyle.Render(line)
		} else if !svc.Enabled {
			line = dimStyle.Render(line)
		} else {
			line = nameStyle(svc.Status).Render(line)
		}

		lines = append(lines, line)
	}

	// Pad remaining lines
	for len(lines) < visible {
		lines = append(lines, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return renderPanel(content, m.width, m.height, "services", focused)
}

func nameStyle(status core.Status) lipgloss.Style {
	switch status {
	case "running", "active":
		return nameRunning
	case "down", "inactive", "disabled":
		return nameDown
	case "failed":
		return nameFailed
	default:
		return nameUnknown
	}
}
