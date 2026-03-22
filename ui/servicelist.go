package ui

import (
	"fmt"
	"lazy-init/core"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	filtered []core.Service
	cursor   int
	offset   int
	width    int
	height   int
	filter   textinput.Model
	filtering bool
}

func newServiceList() serviceListModel {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 64
	return serviceListModel{filter: ti}
}

func (m *serviceListModel) Filtering() bool {
	return m.filtering
}

func (m *serviceListModel) StartFilter() {
	m.filtering = true
	m.filter.SetValue("")
	m.filter.Focus()
	m.applyFilter()
}

func (m *serviceListModel) StopFilter(clear bool) {
	m.filtering = false
	m.filter.Blur()
	if clear {
		m.filter.SetValue("")
	}
	m.applyFilter()
}

func (m *serviceListModel) UpdateFilter(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.applyFilter()
	return cmd
}

func (m *serviceListModel) applyFilter() {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		m.filtered = m.services
	} else {
		m.filtered = nil
		for _, svc := range m.services {
			if fuzzyMatch(strings.ToLower(svc.Name), query) {
				m.filtered = append(m.filtered, svc)
			}
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m.offset = 0
	m.fixOffset()
}

func fuzzyMatch(s, pattern string) bool {
	pi := 0
	for i := 0; i < len(s) && pi < len(pattern); i++ {
		if s[i] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}

func (m *serviceListModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *serviceListModel) Selected() string {
	if len(m.filtered) == 0 {
		return ""
	}
	return m.filtered[m.cursor].Name
}

func (m *serviceListModel) Up() {
	if len(m.filtered) == 0 {
		return
	}
	if m.cursor > 0 {
		m.cursor--
	} else {
		m.cursor = len(m.filtered) - 1
	}
	m.fixOffset()
}

func (m *serviceListModel) Down() {
	if len(m.filtered) == 0 {
		return
	}
	if m.cursor < len(m.filtered)-1 {
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
	if len(m.filtered) == 0 {
		return
	}
	m.cursor = len(m.filtered) - 1
	m.fixOffset()
}

func (m *serviceListModel) SetServices(services []core.Service) {
	m.services = services
	m.applyFilter()
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

	// Reserve a line for the filter input when active or has a query
	filterLine := ""
	if m.filtering || m.filter.Value() != "" {
		m.filter.Width = m.width - 6
		filterLine = "/" + m.filter.View()
		visible--
	}

	var lines []string
	end := m.offset + visible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := m.offset; i < end; i++ {
		svc := m.filtered[i]
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

	if filterLine != "" {
		lines = append(lines, filterLine)
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
