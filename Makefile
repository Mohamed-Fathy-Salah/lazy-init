BINARY  := lazyinit
PREFIX  := /usr/local
BINDIR  := $(PREFIX)/bin
GOFLAGS := -trimpath
LDFLAGS := -s -w

# Detect init system
INIT := $(shell if [ -d /run/runit ] || [ "$(cat /proc/1/comm 2>/dev/null)" = "runit" ]; then echo runit; elif [ -d /run/systemd/system ]; then echo systemd; fi)

.PHONY: all build install uninstall clean

all: build

build:
ifndef INIT
	$(error Could not detect init system. Run: make build INIT=runit or make build INIT=systemd)
endif
	go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -tags $(INIT) -o $(BINARY) ./cmd/lazy-init/

install: build
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)

clean:
	rm -f $(BINARY)
