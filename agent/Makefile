# Local Service Panel Agent Makefile

# Version
VERSION := 0.6.0

# Go module path
MODULE := github.com/user/local-service-panel/agent

# Version ldflag
LDFLAGS := -X '$(MODULE)/internal/version.Version=$(VERSION)'

# Output binary name
BINARY := agent.exe

.PHONY: all build clean

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/agent

clean:
	rm -f $(BINARY)

# Install/Uninstall shortcuts (requires admin)
install:
	$(MAKE) build
	@powershell -Command "Start-Process powershell -Verb RunAs -ArgumentList '-NoProfile -ExecutionPolicy Bypass -File \"$(CURDIR)/../scripts/install.ps1\" -AgentPath \"$(CURDIR)/$(BINARY)\"'"

uninstall:
	@powershell -Command "Start-Process powershell -Verb RunAs -ArgumentList '-NoProfile -ExecutionPolicy Bypass -File \"$(CURDIR)/../scripts/uninstall.ps1\"'"
