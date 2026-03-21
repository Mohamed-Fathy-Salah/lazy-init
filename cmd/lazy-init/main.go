package main

import (
	"fmt"
	"lazy-init/ui"
	"os"
)

func main() {
	m := newManager()
	if err := ui.Run(m); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
