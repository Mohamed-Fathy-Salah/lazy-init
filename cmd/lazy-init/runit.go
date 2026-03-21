//go:build runit

package main

import (
	"lazy-init/adapter/runit"
	"lazy-init/core"
)

func newManager() core.ServiceManager {
	return runit.New()
}