package ui

import (
	"fmt"
	"lazy-init/core"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	valGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	valRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	valYellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	valDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type detailModel struct {
	service *core.Service
	width   int
	height  int
}

func (m *detailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *detailModel) SetService(svc *core.Service) {
	m.service = svc
}

func (m *detailModel) View(focused bool) string {
	if m.height < 3 {
		return ""
	}

	var content string
	if m.service == nil {
		content = valDim.Render("no service selected")
	} else {
		svc := m.service
		lines := []string{
			labelStyle.Render("Name:    ") + svc.Name,
			labelStyle.Render("Status:  ") + statusVal(svc.Status),
			labelStyle.Render("Enabled: ") + enabledVal(svc.Enabled),
		}
		if svc.PID > 0 {
			lines = append(lines, labelStyle.Render("PID:     ")+fmt.Sprintf("%d", svc.PID))
		}
		if svc.Uptime > 0 {
			lines = append(lines, labelStyle.Render("Uptime:  ")+formatUptime(svc.Uptime))
		}
		if svc.Command != "" {
			lines = append(lines, labelStyle.Render("Command: ")+svc.Command)
		}
		if svc.Extra != "" {
			lines = append(lines, labelStyle.Render("Info:    ")+svc.Extra)
		}
		content = fmt.Sprintf("%s", joinLines(lines))
	}

	return renderPanel(content, m.width, m.height, "detail", focused)
}

func statusVal(s core.Status) string {
	switch s {
	case "running", "active":
		return valGreen.Render(string(s))
	case "down", "inactive", "disabled":
		return valRed.Render(string(s))
	case "failed":
		return valRed.Bold(true).Render(string(s))
	default:
		return valYellow.Render(string(s))
	}
}

func enabledVal(enabled bool) string {
	if enabled {
		return valGreen.Render("yes")
	}
	return valRed.Render("no")
}

func formatUptime(d time.Duration) string {
	d = d.Truncate(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
