//go:build runit

package runit

import (
	"fmt"
	"io"
	"lazy-init/core"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const serviceDir = "/var/service"

type manager struct{}

// List implements [core.ServiceManager].
func (m *manager) List() ([]core.Service, error) {
	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, fmt.Errorf("reading service directory: %w", err)
	}

	var services []core.Service
	for _, entry := range entries {
		name := entry.Name()
		status := statusFor(name)
		services = append(services, core.Service{
			Name:   name,
			Status: status,
		})
	}
	return services, nil
}

// Logs implements [core.ServiceManager].
func (m *manager) Logs(name string) (io.Reader, error) {
	logDir := filepath.Join(serviceDir, name, "log", "main")
	currentLog := filepath.Join(logDir, "current")
	f, err := os.Open(currentLog)
	if err != nil {
		return nil, fmt.Errorf("opening log for %s: %w", name, err)
	}
	return f, nil
}

// Start implements [core.ServiceManager].
func (m *manager) Start(name string) error {
	return sv("start", name)
}

// Stop implements [core.ServiceManager].
func (m *manager) Stop(name string) error {
	return sv("stop", name)
}

func sv(command, name string) error {
	cmd := exec.Command("sv", command, filepath.Join(serviceDir, name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func statusFor(name string) core.Status {
	out, err := exec.Command("sv", "status", filepath.Join(serviceDir, name)).CombinedOutput()
	if err != nil {
		return "unknown"
	}
	line := strings.TrimSpace(string(out))
	if strings.HasPrefix(line, "run:") {
		return "running"
	}
	if strings.HasPrefix(line, "down:") {
		return "down"
	}
	return "unknown"
}

func New() core.ServiceManager {
	return &manager{}
}