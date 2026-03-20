# lazy-init

A lazygit-style TUI for managing init system services.

![lazy-init](assets/image1.png)

Services are colored by status: green (running), red (down), dim (disabled).

## Build

```sh
# for runit systems
go build -tags runit -o lazy-init ./cmd/lazy-init/

# for systemd systems
go build -tags systemd -o lazy-init ./cmd/lazy-init/
```

## Usage

```sh
sudo ./lazy-init
```

Root is required to read service status and control services.

## Key Bindings

### Services panel

| Key | Action |
|-----|--------|
| `j` / `↓` | Cursor down (cyclic) |
| `k` / `↑` | Cursor up (cyclic) |
| `g` | Go to first service |
| `G` | Go to last service |
| `enter` | Load logs for selected service |
| `s` | Start service |
| `x` | Stop service |
| `e` | Enable service (start on boot) |
| `d` | Disable service |

### Logs panel

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `g` | Go to top |
| `G` | Go to bottom |

### Global

| Key | Action |
|-----|--------|
| `tab` | Switch panel focus |
| `q` / `ctrl+c` | Quit |

## Detail Panel

The detail panel shows information about the selected service:

- **Name** - service name
- **Status** - current state (running, down, disabled, etc.)
- **Enabled** - whether the service starts on boot
- **PID** - process ID (when running)
- **Uptime** - how long the service has been running
- **Command** - the command the service executes
- **Info** - additional state (e.g. log process info, want state)

The service list and detail panel auto-refresh every 2 seconds. Logs live-tail when scrolled to the bottom.

## Supported Init Systems

- **runit** - reads from `/etc/sv` (available) and `/var/service` (enabled), uses `sv` for control
- **systemd** - uses `systemctl` and `journalctl`

### Adding a new init system

1. Create `adapter/<name>/manager.go` with a build tag, implementing `core.ServiceManager`
2. Create `cmd/lazy-init/<name>.go` with the same build tag, providing `newManager()`
