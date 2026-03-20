package core

import "io"

type Status string

type Service struct {
    Name   string
    Status Status
}

type ServiceManager interface {
    List() ([]Service, error)
    Start(name string) error
    Stop(name string) error
    Logs(name string) (io.Reader, error)
}
