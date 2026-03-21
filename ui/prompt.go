package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type promptKind int

const (
	promptNone promptKind = iota
	promptAddName
	promptConfirmRemove
)

type promptModel struct {
	kind      promptKind
	label     string
	input     textinput.Model
	value     string // stored value from previous step (e.g. name for add)
	width     int
}

func newPrompt() promptModel {
	ti := textinput.New()
	ti.CharLimit = 128
	return promptModel{input: ti}
}

func (m *promptModel) Start(kind promptKind, label, placeholder string) {
	m.kind = kind
	m.label = label
	m.input.SetValue("")
	m.input.Placeholder = placeholder
	m.input.Focus()
}

func (m *promptModel) Active() bool {
	return m.kind != promptNone
}

func (m *promptModel) Cancel() {
	m.kind = promptNone
	m.value = ""
	m.input.Blur()
}

func (m *promptModel) Submit() (promptKind, string) {
	kind := m.kind
	val := m.input.Value()
	m.kind = promptNone
	m.input.Blur()
	return kind, val
}

func (m *promptModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

var (
	promptBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5")).
			Padding(0, 1)
	promptLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
)

func (m *promptModel) View(width int) string {
	if !m.Active() {
		return ""
	}
	m.input.Width = width - 6
	content := promptLabelStyle.Render(m.label) + "\n" + m.input.View()
	return promptBoxStyle.Width(width - 4).Render(content)
}
