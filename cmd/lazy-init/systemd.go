//go:build systemd

package main

import (
	"lazy-init/adapter/systemd"
	"lazy-init/core"
)

func newManager() core.ServiceManager {
	return systemd.New()
}