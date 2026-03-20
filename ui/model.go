package ui

import (
	"io"
	"lazy-init/core"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	listWidth    = 35
	refreshInterval = 2 * time.Second
	maxLogSize   = 64 * 1024
)

type panel int

const (
	panelList panel = iota
	panelLogs
)

// Messages
type servicesLoadedMsg struct {
	services []core.Service
	err      error
}

type logsLoadedMsg struct {
	name    string
	content string
	err     error
}

type tickMsg time.Time

type model struct {
	manager     core.ServiceManager
	serviceList serviceListModel
	logViewer   logViewerModel
	activePanel panel
	width       int
	height      int
}

func newModel(mgr core.ServiceManager) model {
	return model{
		manager:     mgr,
		activePanel: panelList,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadServicesCmd(m.manager), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.serviceList.SetSize(listWidth, m.height)
		m.logViewer.SetSize(m.width-listWidth, m.height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.activePanel == panelList {
				m.activePanel = panelLogs
			} else {
				m.activePanel = panelList
			}
			return m, nil
		}

		if m.activePanel == panelList {
			return m.updateList(msg)
		}
		cmd := m.logViewer.Update(msg)
		return m, cmd

	case servicesLoadedMsg:
		if msg.err == nil {
			m.serviceList.services = msg.services
			if m.serviceList.cursor >= len(msg.services) {
				m.serviceList.cursor = max(0, len(msg.services)-1)
			}
		}
		return m, nil

	case logsLoadedMsg:
		if msg.err == nil {
			m.logViewer.SetContent(msg.name, msg.content)
		} else {
			m.logViewer.SetContent(msg.name, "error: "+msg.err.Error())
		}
		return m, nil

	case tickMsg:
		return m, tea.Batch(loadServicesCmd(m.manager), tickCmd())
	}

	return m, nil
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.serviceList.Down()
	case "k", "up":
		m.serviceList.Up()
	case "enter":
		if name := m.serviceList.Selected(); name != "" {
			return m, loadLogsCmd(m.manager, name)
		}
	case "s":
		if name := m.serviceList.Selected(); name != "" {
			return m, startServiceCmd(m.manager, name)
		}
	case "x":
		if name := m.serviceList.Selected(); name != "" {
			return m, stopServiceCmd(m.manager, name)
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}
	left := m.serviceList.View(m.activePanel == panelList)
	right := m.logViewer.View(m.activePanel == panelLogs)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// Commands

func loadServicesCmd(mgr core.ServiceManager) tea.Cmd {
	return func() tea.Msg {
		services, err := mgr.List()
		return servicesLoadedMsg{services: services, err: err}
	}
}

func loadLogsCmd(mgr core.ServiceManager, name string) tea.Cmd {
	return func() tea.Msg {
		r, err := mgr.Logs(name)
		if err != nil {
			return logsLoadedMsg{name: name, err: err}
		}
		b, err := io.ReadAll(io.LimitReader(r, maxLogSize))
		if err != nil {
			return logsLoadedMsg{name: name, err: err}
		}
		if rc, ok := r.(io.ReadCloser); ok {
			rc.Close()
		}
		return logsLoadedMsg{name: name, content: string(b)}
	}
}

func startServiceCmd(mgr core.ServiceManager, name string) tea.Cmd {
	return func() tea.Msg {
		mgr.Start(name)
		services, err := mgr.List()
		return servicesLoadedMsg{services: services, err: err}
	}
}

func stopServiceCmd(mgr core.ServiceManager, name string) tea.Cmd {
	return func() tea.Msg {
		mgr.Stop(name)
		services, err := mgr.List()
		return servicesLoadedMsg{services: services, err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Run starts the TUI.
func Run(mgr core.ServiceManager) error {
	p := tea.NewProgram(newModel(mgr), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
