package ui

import (
	"io"
	"lazy-init/core"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	leftWidth       = 35
	detailHeight    = 11
	refreshInterval = 2 * time.Second
	maxLogSize      = 64 * 1024
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

// Message returned after editor exits
type editorFinishedMsg struct{ err error }

type model struct {
	manager     core.ServiceManager
	detail      detailModel
	serviceList serviceListModel
	logViewer   logViewerModel
	prompt      promptModel
	activePanel panel
	width       int
	height      int
}

func newModel(mgr core.ServiceManager) model {
	return model{
		manager:     mgr,
		activePanel: panelList,
		prompt:      newPrompt(),
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
		m.resize()
		return m, nil

	case tea.KeyMsg:
		// Handle prompt input first
		if m.prompt.Active() {
			return m.updatePrompt(msg)
		}

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

		switch m.activePanel {
		case panelList:
			return m.updateList(msg)
		case panelLogs:
			cmd := m.logViewer.Update(msg)
			return m, cmd
		}
		return m, nil

	case editorFinishedMsg:
		return m, loadServicesCmd(m.manager)

	case servicesLoadedMsg:
		if msg.err == nil {
			m.serviceList.services = msg.services
			if m.serviceList.cursor >= len(msg.services) {
				m.serviceList.cursor = max(0, len(msg.services)-1)
			}
			m.updateDetail()
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
		cmds := []tea.Cmd{loadServicesCmd(m.manager), tickCmd()}
		if m.logViewer.serviceName != "" {
			cmds = append(cmds, loadLogsCmd(m.manager, m.logViewer.serviceName))
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m *model) resize() {
	panelHeight := m.height - 1 // help bar
	m.detail.SetSize(leftWidth, detailHeight)
	m.serviceList.SetSize(leftWidth, panelHeight-detailHeight)
	m.logViewer.SetSize(m.width-leftWidth, panelHeight)
}

func (m *model) updateDetail() {
	if len(m.serviceList.services) > 0 && m.serviceList.cursor < len(m.serviceList.services) {
		svc := m.serviceList.services[m.serviceList.cursor]
		m.detail.SetService(&svc)
	} else {
		m.detail.SetService(nil)
	}
}

func (m model) updatePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.prompt.Cancel()
		return m, nil
	case "enter":
		kind, val := m.prompt.Submit()
		switch kind {
		case promptAddName:
			if val == "" {
				return m, nil
			}
			return m, createAndEditCmd(m.manager, val)
		case promptConfirmRemove:
			if val == "y" || val == "Y" || val == "yes" {
				name := m.prompt.value
				m.prompt.value = ""
				return m, removeServiceCmd(m.manager, name)
			}
			return m, nil
		}
		return m, nil
	}
	cmd := m.prompt.Update(msg)
	return m, cmd
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.serviceList.Down()
		m.updateDetail()
	case "k", "up":
		m.serviceList.Up()
		m.updateDetail()
	case "g":
		m.serviceList.GoTop()
		m.updateDetail()
	case "G":
		m.serviceList.GoBottom()
		m.updateDetail()
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
	case "e":
		if name := m.serviceList.Selected(); name != "" {
			return m, enableServiceCmd(m.manager, name)
		}
	case "d":
		if name := m.serviceList.Selected(); name != "" {
			return m, disableServiceCmd(m.manager, name)
		}
	case "a":
		m.prompt.Start(promptAddName, "New service name:", "my-service")
		return m, nil
	case "r":
		if name := m.serviceList.Selected(); name != "" {
			m.prompt.value = name
			m.prompt.Start(promptConfirmRemove, "Remove "+name+"? (y/n)", "")
			return m, nil
		}
	case "E":
		if name := m.serviceList.Selected(); name != "" {
			return m, editServiceCmd(m.manager, name)
		}
	}
	return m, nil
}

var (
	helpKeyStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	helpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	detail := m.detail.View(false)
	list := m.serviceList.View(m.activePanel == panelList)
	logs := m.logViewer.View(m.activePanel == panelLogs)

	leftCol := lipgloss.JoinVertical(lipgloss.Left, detail, list)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, logs)

	var help string
	if m.prompt.Active() {
		help = helpKeyStyle.Render("enter") + helpDescStyle.Render(" confirm  ") +
			helpKeyStyle.Render("esc") + helpDescStyle.Render(" cancel")
		return panels + "\n" + m.prompt.View(m.width) + "\n" + help
	}

	if m.activePanel == panelList {
		help = helpKeyStyle.Render("j/k") + helpDescStyle.Render(" navigate  ") +
			helpKeyStyle.Render("g/G") + helpDescStyle.Render(" top/bottom  ") +
			helpKeyStyle.Render("enter") + helpDescStyle.Render(" logs  ") +
			helpKeyStyle.Render("s") + helpDescStyle.Render(" start  ") +
			helpKeyStyle.Render("x") + helpDescStyle.Render(" stop  ") +
			helpKeyStyle.Render("e") + helpDescStyle.Render(" enable  ") +
			helpKeyStyle.Render("d") + helpDescStyle.Render(" disable  ") +
			helpKeyStyle.Render("a") + helpDescStyle.Render(" add  ") +
			helpKeyStyle.Render("r") + helpDescStyle.Render(" remove  ") +
			helpKeyStyle.Render("E") + helpDescStyle.Render(" edit  ")
	} else {
		help = helpKeyStyle.Render("j/k") + helpDescStyle.Render(" scroll  ") +
			helpKeyStyle.Render("g/G") + helpDescStyle.Render(" top/bottom  ")
	}
	help += helpKeyStyle.Render("tab") + helpDescStyle.Render(" switch panel  ") +
		helpKeyStyle.Render("q") + helpDescStyle.Render(" quit")

	return panels + "\n" + help
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

func enableServiceCmd(mgr core.ServiceManager, name string) tea.Cmd {
	return func() tea.Msg {
		mgr.Enable(name)
		services, err := mgr.List()
		return servicesLoadedMsg{services: services, err: err}
	}
}

func disableServiceCmd(mgr core.ServiceManager, name string) tea.Cmd {
	return func() tea.Msg {
		mgr.Disable(name)
		services, err := mgr.List()
		return servicesLoadedMsg{services: services, err: err}
	}
}

func createAndEditCmd(mgr core.ServiceManager, name string) tea.Cmd {
	if err := mgr.Create(name); err != nil {
		return nil
	}
	return editServiceCmd(mgr, name)
}

func removeServiceCmd(mgr core.ServiceManager, name string) tea.Cmd {
	return func() tea.Msg {
		mgr.Remove(name)
		services, err := mgr.List()
		return servicesLoadedMsg{services: services, err: err}
	}
}

func editServiceCmd(mgr core.ServiceManager, name string) tea.Cmd {
	path, err := mgr.EditFile(name)
	if err != nil {
		return nil
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
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
