# lazy-init

A lazygit-style TUI for managing init system services.

```
┌─ services ──────────┬─ logs: sshd ─────────────────┐
│ ● acpid      running│ 2026-03-20 ...                │
│ ● dbus       running│ 2026-03-20 ...                │
│ ▶ sshd       running│ 2026-03-20 ...                │
│ ● docker     running│                               │
│ ● earlyoom     down │                               │
└─────────────────────┴───────────────────────────────┘
```

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

| Key | Action |
|-----|--------|
| `j` / `↓` | Cursor down |
| `k` / `↑` | Cursor up |
| `enter` | Load logs for selected service |
| `s` | Start service |
| `x` | Stop service |
| `tab` | Switch panel focus |
| `pgup` / `pgdn` | Scroll logs |
| `q` / `ctrl+c` | Quit |

## Supported Init Systems

- **runit** — reads from `/var/service`, uses `sv` for control
- **systemd** — uses `systemctl` and `journalctl`

Adding a new init system:

1. Create `adapter/<name>/manager.go` implementing `core.ServiceManager`
2. Create `cmd/lazy-init/<name>.go` with a build tag providing `newManager()`
