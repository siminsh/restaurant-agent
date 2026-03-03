# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**De Gouden Lepel restaurant operations assistant** built in Go. An HTTP server (and interactive TUI) expose a chat endpoint backed by Claude (via the Anthropic Go SDK) that uses tool-calling in an agentic loop for inventory management.

## Build & Run

Go 1.25+. Module: `restaurant-agent`

```bash
# Setup: copy env file and fill in ANTHROPIC_API_KEY
cp .env.example .env

# Build and run HTTP server (port 8080 by default)
make run

# Build and run interactive TUI
make run-tui

# Lint
make vet

# Run tests (none yet, but the target exists)
make test

# Quick manual testing via curl
make health                              # health check
make inventory                           # list all inventory
make chat MSG="What items are low on stock?"  # chat with agent

# Docker
make docker-build && make docker-run
```

## Configuration

All via environment variables (or `.env` file, loaded by `godotenv`):

| Variable | Default | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | *(required)* | Anthropic API key |
| `PORT` | `8080` | HTTP server port |
| `CLAUDE_MODEL` | `claude-sonnet-4-5-20250929` | Claude model to use |
| `MAX_AGENT_ITERATIONS` | `15` | Max tool-call loops per request |

## API Endpoints

- `POST /api/v1/chat` — send `{"message": "...", "session_id": "..."}` (session_id optional, auto-generated if omitted)
- `GET /api/v1/inventory` — list all inventory items
- `DELETE /api/v1/sessions/{sessionID}` — clear a chat session
- `GET /health` — health check

## Architecture

### Data Model

Inventory and menu data live in `internal/store/memory.go`. Seeded on startup with 25 ingredients and 8 Dutch menu items. State resets on restart. Mutated by inventory tools (add_stock, remove_stock, place_order).

### Agentic Loop

The core loop lives in `pkg/agent/agent.go`: sends messages to Claude with tool definitions → executes any tool calls via the registry → feeds results back → repeats until Claude responds with text only (or hits `MaxAgentIterations`).

Sessions are in-memory maps with per-session mutex locking and a 1-hour TTL (background cleanup every 10 minutes).

### Key Packages

- **`cmd/agent`** — HTTP server entrypoint; wires up config, store, tools, LLM client, agent, and HTTP server
- **`cmd/tui`** — BubbleTea interactive terminal UI
- **`pkg/agent`** — Agent struct with session management and the core `Chat()` agentic loop
- **`pkg/llm`** — Anthropic SDK client wrapper, `GenerateSchema[T]()` for tool schemas (via `invopop/jsonschema` reflection), `SystemPrompt` const, and `BuildToolDefinitions()` defining all 8 tools
- **`pkg/tools`** — `Registry` (name→handler map), `RegisterInventoryTools()` for local store tools
- **`pkg/api`** — chi v5 router, CORS middleware, HTTP handlers
- **`pkg/config`** — loads config from env vars via `godotenv` with defaults
- **`internal/store`** — `MemoryStore` with mutex-protected inventory, menu items, and orders

## Adding a New Tool

1. Define the input struct with `json` and `jsonschema` tags in the appropriate file under `pkg/tools/`
2. Register the handler via `reg.Register()` in the corresponding `Register*Tools()` function
3. Add the corresponding `anthropic.ToolParam` entry in `BuildToolDefinitions()` in `pkg/llm/client.go` — use `GenerateSchema[YourInputStruct]()` for the schema
4. Update the system prompt in `pkg/llm/client.go` if the tool introduces a new capability

**Critical:** Tool names must match exactly between `BuildToolDefinitions()` in `pkg/llm/client.go` and `reg.Register()` calls in `pkg/tools/`. These files are the tool "contract" and must stay in sync.

## Key Conventions

- Tool input schemas are generated from Go structs via `invopop/jsonschema` reflection — annotate fields with `jsonschema:"required,description=..."` tags
- Tool handlers return JSON strings (not Go structs); errors from the store are returned as non-error tool results (e.g., `"Error: ..."`) so Claude can handle them gracefully
- All store map keys (inventory items and menu items) are lowercased via `strings.ToLower`
- The store is seeded with 25 ingredients and 8 menu items matching De Gouden Lepel; state resets on restart
- `AddStock` auto-creates new inventory items if the key doesn't exist; all other operations require existing items
- Key dependencies: `anthropics/anthropic-sdk-go` v1.26.0, `go-chi/chi` v5, `charmbracelet/bubbletea` for TUI
