//go:build runit

package runit

import (
	"fmt"
	"io"
	"lazy-init/core"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	serviceDir  = "/var/service"
	availableDir = "/etc/sv"
)

type manager struct{}

// List implements [core.ServiceManager].
func (m *manager) List() ([]core.Service, error) {
	// Collect enabled services (in /var/service/)
	enabled := map[string]bool{}
	enabledEntries, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, fmt.Errorf("reading service directory: %w", err)
	}
	for _, e := range enabledEntries {
		enabled[e.Name()] = true
	}

	// Collect all available services (in /etc/sv/)
	allEntries, err := os.ReadDir(availableDir)
	if err != nil {
		return nil, fmt.Errorf("reading available services: %w", err)
	}

	var services []core.Service
	for _, entry := range allEntries {
		name := entry.Name()
		svc := core.Service{
			Name:    name,
			Enabled: enabled[name],
		}
		if svc.Enabled {
			parseStatus(&svc)
		} else {
			svc.Status = "disabled"
		}
		svc.Command = readCommand(name)
		services = append(services, svc)
	}
	return services, nil
}

// Logs implements [core.ServiceManager].
func (m *manager) Logs(name string) (io.Reader, error) {
	paths := []string{
		filepath.Join(serviceDir, name, "log", "main", "current"),       // svlogd direct
		filepath.Join("/var/log/socklog/everything", "current"),          // socklog everything
		filepath.Join("/var/log/socklog/daemon", "current"),              // socklog daemon facility
		filepath.Join("/var/log", name),                                  // /var/log/<service>
		filepath.Join("/var/log", name+".log"),                           // /var/log/<service>.log
	}
	for _, p := range paths {
		f, err := os.Open(p)
		if err == nil {
			return f, nil
		}
	}
	return nil, fmt.Errorf("no logs found for %s\n\nservices use vlogger which sends to syslog.\ninstall socklog-void to enable log collection:\n\n  sudo xbps-install -S socklog-void\n  sudo ln -s /etc/sv/socklog-unix /var/service/\n  sudo ln -s /etc/sv/nanoklogd /var/service/", name)
}

// Start implements [core.ServiceManager].
func (m *manager) Start(name string) error {
	return sv("start", name)
}

// Stop implements [core.ServiceManager].
func (m *manager) Stop(name string) error {
	return sv("stop", name)
}

// Enable implements [core.ServiceManager].
func (m *manager) Enable(name string) error {
	src := filepath.Join(availableDir, name)
	dst := filepath.Join(serviceDir, name)
	return os.Symlink(src, dst)
}

// Disable implements [core.ServiceManager].
func (m *manager) Disable(name string) error {
	return os.Remove(filepath.Join(serviceDir, name))
}

func sv(command, name string) error {
	return exec.Command("sv", command, filepath.Join(serviceDir, name)).Run()
}

// parseStatus parses "sv status" output like:
//   run: /var/service/sshd: (pid 763) 3156s; run: log: (pid 761) 3156s
//   down: /var/service/dunst: 1s, normally up, want up
func parseStatus(svc *core.Service) {
	out, err := exec.Command("sv", "status", filepath.Join(serviceDir, svc.Name)).CombinedOutput()
	if err != nil {
		svc.Status = "unknown"
		return
	}
	line := strings.TrimSpace(string(out))

	if strings.HasPrefix(line, "run:") {
		svc.Status = "running"
	} else if strings.HasPrefix(line, "down:") {
		svc.Status = "down"
	} else {
		svc.Status = "unknown"
	}

	// Parse PID: (pid 763)
	if start := strings.Index(line, "(pid "); start != -1 {
		rest := line[start+5:]
		if end := strings.Index(rest, ")"); end != -1 {
			if pid, err := strconv.Atoi(rest[:end]); err == nil {
				svc.PID = pid
			}
		}
	}

	// Parse uptime: number followed by 's' after pid or after colon for down
	if start := strings.Index(line, ") "); start != -1 {
		rest := line[start+2:]
		if end := strings.Index(rest, "s"); end != -1 {
			if secs, err := strconv.Atoi(rest[:end]); err == nil {
				svc.Uptime = time.Duration(secs) * time.Second
			}
		}
	} else {
		// down services: "down: /var/service/x: 1s, ..."
		parts := strings.SplitN(line, ": ", 3)
		if len(parts) == 3 {
			rest := parts[2]
			if end := strings.Index(rest, "s"); end != -1 {
				if secs, err := strconv.Atoi(rest[:end]); err == nil {
					svc.Uptime = time.Duration(secs) * time.Second
				}
			}
		}
	}

	// Parse extra info after semicolon or comma
	if idx := strings.Index(line, ", "); idx != -1 {
		svc.Extra = strings.TrimSpace(line[idx+2:])
	} else if idx := strings.Index(line, "; "); idx != -1 {
		svc.Extra = strings.TrimSpace(line[idx+2:])
	}
}

func readCommand(name string) string {
	runFile := filepath.Join(availableDir, name, "run")
	data, err := os.ReadFile(runFile)
	if err != nil {
		return ""
	}
	// Find the exec line
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "exec ") {
			return strings.TrimPrefix(line, "exec ")
		}
	}
	return ""
}

// Create implements [core.ServiceManager].
func (m *manager) Create(name string) error {
	svcDir := filepath.Join(availableDir, name)
	logDir := filepath.Join(svcDir, "log")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("creating service directory: %w", err)
	}

	runScript := fmt.Sprintf("#!/bin/sh\nexec 2>&1\nexec chpst -u root %s\n", name)
	if err := os.WriteFile(filepath.Join(svcDir, "run"), []byte(runScript), 0755); err != nil {
		return fmt.Errorf("writing run script: %w", err)
	}

	logScript := fmt.Sprintf("#!/bin/sh\nexec vlogger -t %s -p daemon\n", name)
	if err := os.WriteFile(filepath.Join(logDir, "run"), []byte(logScript), 0755); err != nil {
		return fmt.Errorf("writing log script: %w", err)
	}

	return nil
}

// Remove implements [core.ServiceManager].
func (m *manager) Remove(name string) error {
	// Disable first if enabled
	link := filepath.Join(serviceDir, name)
	if _, err := os.Lstat(link); err == nil {
		// Stop the service before removing
		_ = sv("stop", name)
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("disabling service: %w", err)
		}
	}

	svcDir := filepath.Join(availableDir, name)
	if err := os.RemoveAll(svcDir); err != nil {
		return fmt.Errorf("removing service directory: %w", err)
	}
	return nil
}

// EditFile implements [core.ServiceManager].
func (m *manager) EditFile(name string) (string, error) {
	path := filepath.Join(availableDir, name, "run")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("service file not found: %w", err)
	}
	return path, nil
}

func New() core.ServiceManager {
	return &manager{}
}