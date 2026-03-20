package core

import (
    "io"
    "time"
)

type Status string

type Service struct {
    Name    string
    Status  Status
    Enabled bool
    PID     int
    Uptime  time.Duration
    Command string
    Extra   string // additional state info (e.g. "normally up, want up")
}

type ServiceManager interface {
    List() ([]Service, error)
    Start(name string) error
    Stop(name string) error
    Enable(name string) error
    Disable(name string) error
    Logs(name string) (io.Reader, error)
}
