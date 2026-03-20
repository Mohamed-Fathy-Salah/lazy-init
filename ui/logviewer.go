package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type logViewerModel struct {
	viewport    viewport.Model
	serviceName string
	width       int
	height      int
}

func (m *logViewerModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.viewport.Width = w - 2  // border
	m.viewport.Height = h - 2 // border
}

func (m *logViewerModel) SetContent(name, content string) {
	m.serviceName = name
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m *logViewerModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return cmd
}

func (m *logViewerModel) View(focused bool) string {
	border := lipgloss.NormalBorder()
	style := lipgloss.NewStyle().
		Border(border).
		Width(m.width - 2).
		Height(m.height - 2).
		BorderForeground(lipgloss.Color("8"))

	if focused {
		style = style.BorderForeground(lipgloss.Color("4"))
	}

	return style.Render(m.viewport.View())
}
