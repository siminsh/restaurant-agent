.PHONY: build build-tui run run-tui run-race clean vet test chat health inventory docker-build docker-run

# Binary names
BINARY=agent
TUI_BINARY=tui

# Default port
PORT?=8080

# Build the agent binary
build:
	go build -o $(BINARY) ./cmd/agent

# Build the TUI binary
build-tui:
	go build -o $(TUI_BINARY) ./cmd/tui

# Build with race detector enabled
build-race:
	go build -race -o $(BINARY) ./cmd/agent

# Run the agent (requires ANTHROPIC_API_KEY env var)
run: build
	PORT=$(PORT) ./$(BINARY)

# Run the TUI
run-tui: build-tui
	./$(TUI_BINARY)

# Run with race detector (for development)
run-race: build-race
	PORT=$(PORT) ./$(BINARY)

# Run go vet
vet:
	go vet ./...

# Run tests (when tests exist)
test:
	go test ./...

# Run tests with race detector
test-race:
	go test -race ./...

# Clean build artifacts
clean:
	rm -f $(BINARY) $(TUI_BINARY)

# Docker build
docker-build:
	docker build -t restaurant-agent .

# Docker run
docker-run:
	docker run -p 8080:8080 --env-file .env restaurant-agent

# --- Quick curl commands for manual testing ---

# Health check
health:
	@curl -s http://localhost:$(PORT)/health | python3 -m json.tool

# List inventory
inventory:
	@curl -s http://localhost:$(PORT)/api/v1/inventory | python3 -m json.tool

# Chat with the agent (usage: make chat MSG="Give me my morning briefing")
MSG?=Give me my morning briefing
chat:
	@curl -s -X POST http://localhost:$(PORT)/api/v1/chat \
		-H "Content-Type: application/json" \
		-d '{"message": "$(MSG)"}' | python3 -m json.tool
