# Restaurant Operations Agent

An AI-powered restaurant operations agent. Uses Claude to manage inventory through a conversational interface with tool-calling in an agentic loop.

Built in Go.

## What It Does

A restaurant manager opens the agent and asks: *"What items are low on stock?"*

The agent checks inventory levels and responds:

> "You have 13 items below their reorder thresholds. The most critical include biefstuk (6.0 kg, threshold 8.0), mosselen (8.0 kg, threshold 10.0), and tomaten (3.0 kg, threshold 5.0). Shall I place restock orders for the most critical items?"

## Features

### Inventory Management
- Check stock levels, record deliveries, track waste/spoilage
- Low-stock alerts with reorder threshold monitoring
- Menu feasibility checks (can we make 20 biefstukken tonight?)
- Place restock orders with suppliers
- Full inventory reports by category

## Quick Start

```bash
# Set your API key
cp .env.example .env
# Edit .env with your ANTHROPIC_API_KEY

# Build and run the HTTP server
make run

# Or run the interactive TUI
make run-tui
```

The server starts on `http://localhost:8080` by default.

## Configuration

All configuration is via environment variables (or `.env` file):

| Variable | Default | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | *(required)* | Anthropic API key |
| `PORT` | `8080` | HTTP server port |
| `CLAUDE_MODEL` | `claude-sonnet-4-5-20250929` | Claude model to use |
| `MAX_AGENT_ITERATIONS` | `15` | Max tool-call loops per request |

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/chat` | Chat with the agent |
| `GET` | `/api/v1/inventory` | List all inventory items |
| `DELETE` | `/api/v1/sessions/{id}` | Clear a chat session |
| `GET` | `/health` | Health check |

### Chat Request

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What items are low on stock?", "session_id": "optional-session-id"}'
```

## Demo Scenarios

```bash
# 1. Inventory check
make chat MSG="What items are low on stock?"

# 2. Menu feasibility
make chat MSG="Can we serve 30 stamppot boerenkool for tonight's event?"

# 3. Place restock orders
make chat MSG="Place orders for everything that needs restocking."

# 4. Inventory report
make chat MSG="Give me a full inventory report for the produce category."

# 5. Record a delivery
make chat MSG="We just received 10 kg of zalm and 5 kg of biefstuk."
```

## Tools

The agent exposes 8 tools to Claude:

| Tool | Description |
|---|---|
| `check_inventory` | Check stock of a specific ingredient |
| `add_stock` | Record an incoming delivery |
| `remove_stock` | Record waste, spoilage, or corrections |
| `list_low_stock` | List items below reorder thresholds |
| `check_menu_feasibility` | Check if a dish can be made for N servings |
| `place_order` | Place a restock order |
| `get_inventory_report` | Full or filtered inventory snapshot |
| `get_menu_items` | List menu items and ingredients |

## Architecture

```
cmd/agent/         — HTTP server entrypoint
cmd/tui/           — Interactive terminal UI (BubbleTea)
pkg/agent/         — Agentic loop and session management
pkg/llm/           — Anthropic SDK client, tool definitions, system prompt
pkg/tools/         — Tool registry and inventory tool handlers
pkg/api/           — HTTP router, CORS middleware, handlers
pkg/config/        — Environment-based configuration
internal/store/    — In-memory inventory, menu, and order store
```

The agentic loop (`pkg/agent/agent.go`) sends messages to Claude with tool definitions, executes tool calls via the registry, feeds results back, and repeats until Claude responds with text only.

## Docker

```bash
docker build -t restaurant-agent .
docker run -p 8080:8080 --env-file .env restaurant-agent
```

## Makefile Commands

```bash
make build        # Build the HTTP server
make build-tui    # Build the TUI
make build-race   # Build with race detector
make run          # Build and run HTTP server
make run-tui      # Build and run TUI
make run-race     # Run with race detector
make vet          # Run go vet
make test         # Run tests
make test-race    # Run tests with race detector
make clean        # Remove build artifacts
make docker-build # Build Docker image
make docker-run   # Run Docker container
make health       # curl health endpoint
make inventory    # curl inventory endpoint
make chat MSG="..." # Send a chat message
```
