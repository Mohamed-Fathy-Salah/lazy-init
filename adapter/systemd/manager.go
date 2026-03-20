//go:build systemd

package systemd

import (
	"encoding/json"
	"fmt"
	"io"
	"lazy-init/core"
	"os/exec"
	"strings"
)

type manager struct{}

type unitInfo struct {
	Unit   string `json:"unit"`
	Active string `json:"active"`
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
		services = append(services, core.Service{
			Name:   strings.TrimSuffix(u.Unit, ".service"),
			Status: core.Status(u.Active),
		})
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

func New() core.ServiceManager {
	return &manager{}
}