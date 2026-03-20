//go:build systemd

package systemd

import (
	"encoding/json"
	"fmt"
	"io"
	"lazy-init/core"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type manager struct{}

type unitInfo struct {
	Unit      string `json:"unit"`
	Active    string `json:"active"`
	UnitState string `json:"unit_file_state"`
}

// List implements [core.ServiceManager].
func (m *manager) List() ([]core.Service, error) {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--output=json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing systemd units: %w", err)
	}

	var units []unitInfo
	if err := json.Unmarshal(out, &units); err != nil {
		return nil, fmt.Errorf("parsing systemd output: %w", err)
	}

	var services []core.Service
	for _, u := range units {
		name := strings.TrimSuffix(u.Unit, ".service")
		svc := core.Service{
			Name:    name,
			Status:  core.Status(u.Active),
			Enabled: u.UnitState == "enabled",
		}
		parseProps(&svc)
		services = append(services, svc)
	}
	return services, nil
}

// Logs implements [core.ServiceManager].
func (m *manager) Logs(name string) (io.Reader, error) {
	cmd := exec.Command("journalctl", "-u", name, "--no-pager")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating pipe for %s logs: %w", name, err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting journalctl for %s: %w", name, err)
	}
	return stdout, nil
}

// Start implements [core.ServiceManager].
func (m *manager) Start(name string) error {
	return exec.Command("systemctl", "start", name).Run()
}

// Stop implements [core.ServiceManager].
func (m *manager) Stop(name string) error {
	return exec.Command("systemctl", "stop", name).Run()
}

// Enable implements [core.ServiceManager].
func (m *manager) Enable(name string) error {
	return exec.Command("systemctl", "enable", name).Run()
}

// Disable implements [core.ServiceManager].
func (m *manager) Disable(name string) error {
	return exec.Command("systemctl", "disable", name).Run()
}

func parseProps(svc *core.Service) {
	unit := svc.Name + ".service"
	out, err := exec.Command("systemctl", "show", unit,
		"--property=MainPID,ActiveEnterTimestampMonotonic,ExecStart,StatusText",
		"--no-pager").Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch k {
		case "MainPID":
			if pid, err := strconv.Atoi(v); err == nil {
				svc.PID = pid
			}
		case "ActiveEnterTimestampMonotonic":
			if usec, err := strconv.ParseInt(v, 10, 64); err == nil && usec > 0 {
				svc.Uptime = time.Duration(usec) * time.Microsecond
			}
		case "ExecStart":
			// Format: { path=/usr/bin/foo ; argv[]=/usr/bin/foo arg1 ... }
			if start := strings.Index(v, "argv[]="); start != -1 {
				rest := v[start+7:]
				if end := strings.Index(rest, " ;"); end != -1 {
					svc.Command = rest[:end]
				} else {
					rest = strings.TrimRight(rest, " }")
					svc.Command = rest
				}
			}
		case "StatusText":
			if v != "" {
				svc.Extra = v
			}
		}
	}
}

func New() core.ServiceManager {
	return &manager{}
}