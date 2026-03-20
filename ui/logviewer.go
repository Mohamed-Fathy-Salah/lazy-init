package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	newService := m.serviceName != name
	m.serviceName = name
	atBottom := m.viewport.AtBottom()
	m.viewport.SetContent(content)
	if newService || atBottom {
		m.viewport.GotoBottom()
	}
}

func (m *logViewerModel) Update(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "g":
			m.viewport.GotoTop()
			return nil
		case "G":
			m.viewport.GotoBottom()
			return nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return cmd
}

func (m *logViewerModel) View(focused bool) string {
	title := "logs"
	if m.serviceName != "" {
		title = "logs: " + m.serviceName
	}
	return renderPanel(m.viewport.View(), m.width, m.height, title, focused)
}
