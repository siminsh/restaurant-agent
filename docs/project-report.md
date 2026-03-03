# De Gouden Lepel - Restaurant Operations Agent

## Project Report — Prosus/Toqan Hackathon

---

## 1. Project Goal

### The Problem

Restaurant managers juggle three critical operational domains every day: **inventory**, **orders/revenue**, and **workforce scheduling**. These are typically managed in separate systems, making it hard to connect insights across domains. For example, a manager might not realize that their top-selling dish depends on an ingredient that's about to run out, or that they're overstaffed on a slow day.

### The Solution

We built an **AI-powered operations assistant** for De Gouden Lepel, a Dutch restaurant in Amsterdam. The agent provides a single conversational interface that combines all three domains into unified, actionable intelligence.

A restaurant manager can ask:

> "Give me my morning briefing"

And the agent responds with a complete operational overview: today's revenue, top sellers this week, staff schedule, absences, labor costs, and inventory alerts — all connected. If a top-selling item depends on a low-stock ingredient, the agent flags it immediately and suggests a specific restock order.

### Why This Matters for Toqan

Toqan is a platform for deploying AI agents to restaurants, serving 10,000+ daily users. This agent demonstrates how cross-domain intelligence can solve real operational problems for actual restaurants — not just answer simple questions, but reason across multiple data sources and suggest specific actions.

---

## 2. Implementation Approach

### Technology Stack

| Component        | Technology                        |
|------------------|-----------------------------------|
| Language         | Go 1.25                           |
| AI Model         | Claude (via Anthropic Go SDK)     |
| HTTP Framework   | chi v5                            |
| Terminal UI      | BubbleTea (Charm)                 |
| Analytics API    | Toqan Cube.js API                 |
| Containerization | Docker (multi-stage Alpine build) |

### Design Decisions

**Why Go?** Go gives us a single compiled binary with no runtime dependencies, fast startup time, and built-in concurrency. This is ideal for a server that makes parallel API calls and needs to be containerized for the Toqan platform.

**Why an agentic loop (not a simple chatbot)?** Restaurant queries are inherently multi-step. A "morning briefing" requires fetching order data, top sellers, staff schedules, labor costs, absences, and inventory alerts — then synthesizing them. An agentic loop lets Claude decide which tools to call and in what order, handling complex queries without us hard-coding the logic.

**Why tool-calling instead of RAG?** The data is structured and operational (numbers, schedules, stock levels). Tool-calling gives Claude precise, up-to-date data from APIs and the inventory store, rather than searching through documents.

**Why in-memory inventory?** For the hackathon, in-memory state with seed data lets us demonstrate the full inventory workflow (check stock, add deliveries, record waste, place orders) without needing a database. The Toqan platform guide includes a Google Sheets alternative for persistence.

---

## 3. Architecture

### High-Level Architecture Diagram

```
+------------------+       +------------------+       +------------------+
|                  |       |                  |       |                  |
|   HTTP Client    |       |   Terminal (TUI) |       |  Toqan Platform  |
|   (curl / app)   |       |   (BubbleTea)    |       |  (future deploy) |
|                  |       |                  |       |                  |
+--------+---------+       +--------+---------+       +--------+---------+
         |                          |                          |
         v                          v                          v
+--------+----------+------+--------+----------+------+--------+----------+
|                                                                         |
|                        HTTP Server (chi v5)                             |
|                     POST /api/v1/chat                                   |
|                     GET  /api/v1/inventory                              |
|                     DELETE /api/v1/sessions/{id}                        |
|                     GET  /health                                        |
|                                                                         |
+---------+-----------------------------------------------------------+---+
          |                                                           |
          v                                                           v
+---------+---------+                                    +------------+---+
|                   |                                    |                |
|   Agent           |                                    |  Inventory API |
|   (Agentic Loop)  |                                    |  Handler       |
|                   |                                    |                |
|  Session mgmt     |                                    +--------+-------+
|  Conversation     |                                             |
|  history          |                                             v
|                   |                                    +--------+-------+
+---------+---------+                                    |                |
          |                                              |  MemoryStore   |
          v                                              |  (in-memory)   |
+---------+---------+                                    |                |
|                   |                                    +----------------+
|  Claude API       |
|  (Anthropic SDK)  |             TOOL EXECUTION
|                   |        +------------------------+
|  - Send messages  | -----> |                        |
|  - Receive tool   |        |   Tool Registry        |
|    call requests  |        |                        |
|  - Feed results   | <----- |   13 tools registered  |
|    back           |        |                        |
|                   |        +-----+----------+-------+
+-------------------+              |          |
                                   v          v
                          +--------+--+  +----+----------+
                          |           |  |               |
                          | Inventory |  | Operations    |
                          | Tools (8) |  | Tools (5)     |
                          |           |  |               |
                          +-----+-----+  +-------+-------+
                                |                |
                                v                v
                          +-----+-----+  +-------+-------+
                          |           |  |               |
                          | Memory    |  | Toqan API     |
                          | Store     |  | (Cube.js)     |
                          | (local)   |  | (remote)      |
                          |           |  |               |
                          +-----------+  +---------------+
```

### The Agentic Loop

This is the core of the system. When a user sends a message:

```
1. User message is added to session history
2. LOOP (up to MaxAgentIterations = 15):
   a. Send full conversation history + tool definitions to Claude
   b. Claude responds with either:
      - Text only     --> Return response to user (DONE)
      - Tool calls    --> Execute each tool via the Registry
   c. Tool results are added to history as a user message
   d. Go back to step (a)
3. If max iterations reached, return error
```

This allows Claude to chain multiple tool calls. For example, a "morning briefing" triggers `get_daily_briefing` (which makes 5 parallel API calls), and Claude might then follow up with `list_low_stock` to check specific items it noticed are relevant.

### Dual Data Model

The agent works with two fundamentally different data sources:

| Aspect          | Inventory (Local)              | Operations (Remote)                |
|-----------------|--------------------------------|------------------------------------|
| Data store      | In-memory Go maps              | Toqan Cube.js API                  |
| Mutability      | Read + Write                   | Read-only                          |
| Persistence     | Resets on restart              | Persistent (external service)      |
| Data types      | Stock levels, menu items       | Orders, revenue, shifts, absences  |
| Tool count      | 8 tools                        | 5 tools                            |
| Concurrency     | Mutex-protected                | HTTP calls (15s timeout)           |

---

## 4. File Structure and Packages

```
toqan-hackathon/
|-- cmd/
|   |-- agent/
|   |   +-- main.go              # HTTP server entrypoint
|   +-- tui/
|       |-- main.go              # TUI entrypoint
|       +-- model.go             # BubbleTea model, views, styles
|
|-- pkg/
|   |-- agent/
|   |   +-- agent.go             # Agentic loop, session management
|   |-- llm/
|   |   +-- client.go            # Anthropic SDK wrapper, tool definitions,
|   |                            # schema generation, system prompt
|   |-- tools/
|   |   |-- registry.go          # Tool name -> handler function map
|   |   |-- inventory.go         # 8 inventory tool handlers
|   |   +-- operations.go        # 5 operations tool handlers (Toqan API)
|   |-- api/
|   |   |-- router.go            # chi router, CORS middleware
|   |   +-- handlers.go          # HTTP handlers for chat, inventory, etc.
|   |-- config/
|   |   +-- config.go            # Environment variable configuration
|   +-- toqan/
|       +-- client.go            # HTTP client for Toqan Cube.js API
|
|-- internal/
|   +-- store/
|       +-- memory.go            # In-memory store with seed data
|
|-- docs/
|   +-- toqan-setup-guide.md     # Toqan platform deployment guide
|
|-- Dockerfile                   # Multi-stage Alpine build
|-- Makefile                     # Build, run, test, curl shortcuts
|-- go.mod / go.sum              # Go module dependencies
|-- .env.example                 # Environment variable template
+-- CLAUDE.md                    # AI coding assistant instructions
```

### Package Responsibilities

**`cmd/agent/main.go`** — The HTTP server entrypoint. Creates and wires together all components in order: Config -> MemoryStore -> ToqanClient -> ToolRegistry -> LLMClient -> Agent -> HTTPRouter. Handles graceful shutdown via OS signals.

**`cmd/tui/`** — An interactive terminal UI built with BubbleTea. Provides a chat interface with styled message bubbles, a spinner while the agent is thinking, and markdown rendering of agent responses via Glamour.

**`pkg/agent/agent.go`** — Contains the `Agent` struct and the core `Chat()` method that implements the agentic loop. Manages per-session conversation history in a `map[string][]MessageParam` with per-session mutex locking for concurrent safety. Sessions expire after 1 hour with a background cleanup goroutine.

**`pkg/llm/client.go`** — Wraps the Anthropic Go SDK. Three critical pieces:
1. `GenerateSchema[T]()` — Uses `invopop/jsonschema` reflection to generate JSON schemas from Go structs for tool input definitions.
2. `BuildToolDefinitions()` — Defines all 13 tools with names, descriptions, and schemas. This is one half of the tool "contract".
3. `SystemPrompt` — The full system prompt that instructs Claude on its role, capabilities, and guidelines.

**`pkg/tools/registry.go`** — A simple `map[string]ToolFunc` that maps tool names to handler functions. The `Execute()` method looks up and calls the right handler. This is the other half of the tool "contract".

**`pkg/tools/inventory.go`** — Registers 8 inventory tools. Each tool handler unmarshals the JSON input into a typed Go struct, calls the appropriate `MemoryStore` method, and returns a JSON string result.

**`pkg/tools/operations.go`** — Registers 5 operations tools that query the external Toqan API. The `get_daily_briefing` tool is notable: it launches 5 goroutines to make parallel Cube.js queries (orders, top items, shifts, labor costs, absences), collects results via a channel, and appends local low-stock alerts.

**`pkg/toqan/client.go`** — HTTP client for the Toqan Cube.js analytics API. Sends `POST /cubejs-api/v1/load` requests with structured queries (measures, dimensions, time ranges, filters). Includes `PeriodToDateRange()` which converts human periods ("today", "this_week") to date range arrays.

**`pkg/api/`** — HTTP routing via chi v5 with middleware (request ID, logging, recovery, 120s timeout, CORS). Handlers are thin: they parse the request, call the agent or store, and return JSON.

**`pkg/config/config.go`** — Loads configuration from environment variables with sensible defaults. Uses `godotenv` to load `.env` files automatically.

**`internal/store/memory.go`** — The `MemoryStore` holds three data structures behind a `sync.RWMutex`: inventory items (25 ingredients), menu items (8 dishes with ingredient recipes), and orders (placed restock orders). All map keys are lowercased. The `seed()` method populates with realistic De Gouden Lepel data.

---

## 5. API Reference

### POST /api/v1/chat

Send a message to the agent and receive a response.

**Request:**
```json
{
  "message": "Give me my morning briefing",
  "session_id": "optional-session-id"
}
```

**Response:**
```json
{
  "response": "Here is your morning briefing for today...",
  "session_id": "auto-generated-uuid-if-not-provided",
  "tool_calls": 3
}
```

The `session_id` maintains conversation context across requests. If omitted, a new UUID is generated. The `tool_calls` field indicates how many tools the agent invoked.

### GET /api/v1/inventory

Returns all inventory items with counts.

**Response:**
```json
{
  "items": [
    {
      "name": "biefstuk",
      "quantity": 6,
      "unit": "kg",
      "reorder_threshold": 8,
      "category": "meat",
      "supplier": "Slagerij de Wit",
      "last_updated": "2026-03-02T10:00:00Z"
    }
  ],
  "count": 25
}
```

### DELETE /api/v1/sessions/{sessionID}

Clears a chat session's conversation history.

### GET /health

Returns `{"status": "ok"}`.

---

## 6. Tool Definitions (13 Tools)

### Inventory Tools (8) — Local Data

| Tool | Input | Description |
|------|-------|-------------|
| `check_inventory` | `item_name` | Check stock level of a specific ingredient |
| `add_stock` | `item_name`, `quantity`, `unit` | Record a delivery (auto-creates new items) |
| `remove_stock` | `item_name`, `quantity`, `reason` | Record waste, spoilage, or corrections |
| `list_low_stock` | *(none)* | List items at or below reorder threshold |
| `check_menu_feasibility` | `menu_item`, `servings` | Can we make N servings of this dish? |
| `place_order` | `item_name`, `quantity` | Place a restock order with the supplier |
| `get_inventory_report` | `category` (optional) | Full or filtered inventory snapshot |
| `get_menu_items` | `category` (optional) | List menu items and their ingredients |

### Operations Tools (5) — Live Remote Data

| Tool | Input | Description |
|------|-------|-------------|
| `get_order_summary` | `period` | Order count, revenue, avg value, tips, discounts |
| `get_top_selling_items` | `period`, `limit` | Top menu items by revenue |
| `get_revenue_by_channel` | `period` | Revenue breakdown: dine-in, delivery, takeaway |
| `get_staff_overview` | `period` | Staff schedule, shift count, hours by department |
| `get_daily_briefing` | *(none)* | Comprehensive briefing: 5 parallel API calls + inventory alerts |

**Period values:** `today`, `yesterday`, `this_week`, `last_7_days`, `this_month`, `last_30_days`

---

## 7. Services and External Integrations

### Claude API (Anthropic)

The agent uses Claude as its reasoning engine via the Anthropic Go SDK (`anthropics/anthropic-sdk-go` v1.26.0). Each chat request sends the full conversation history plus 13 tool definitions. Claude decides which tools to call based on the user's question.

**SDK usage pattern:**
```go
resp, err := client.API.Messages.New(ctx, anthropic.MessageNewParams{
    Model:     anthropic.Model("claude-sonnet-4-5-20250929"),
    MaxTokens: 4096,
    System:    []anthropic.TextBlockParam{{Text: llm.SystemPrompt}},
    Messages:  history,
    Tools:     toolDefinitions,
})
```

### Toqan Cube.js API

The Toqan analytics API provides restaurant operational data in Cube.js format. All queries go to `POST /cubejs-api/v1/load` with an API key header.

**Available Cube.js cubes:**
- `Orders` — order count, revenue, tips, discounts, cancellations
- `Revenue` — per-item revenue and quantity sold
- `Shifts` — staff schedules, hours, departments
- `Costs` — labor costs by department
- `Absences` — employee absences with approval status

**Example query:**
```json
{
  "query": {
    "measures": ["Orders.count", "Orders.totalRevenue"],
    "timeDimensions": [{
      "dimension": "Orders.orderedAt",
      "dateRange": ["2026-03-02", "2026-03-02"]
    }]
  }
}
```

---

## 8. Concurrency and Session Design

### Session Management

Each chat session is identified by a UUID and maintains its own conversation history. Key design properties:

- **Per-session locking:** A `sync.Mutex` per session ensures that concurrent requests to the same session are serialized (preventing interleaved tool calls).
- **Global map locking:** A `sync.RWMutex` protects the session map itself for safe concurrent creation and lookup.
- **TTL expiration:** Sessions idle for more than 1 hour are automatically cleaned up by a background goroutine that runs every 10 minutes.

### Parallel API Calls

The `get_daily_briefing` tool demonstrates Go's concurrency model: it launches 5 goroutines (one per Cube.js query), collects results via a buffered channel, and waits for all to complete using a `sync.WaitGroup`. This reduces latency from 5 sequential API calls to roughly the time of the slowest single call.

---

## 9. Deployment

### Docker

The Dockerfile uses a multi-stage build:
1. **Build stage:** Go 1.25 Alpine compiles a static binary with `CGO_ENABLED=0`
2. **Runtime stage:** Alpine 3.21 with just the binary and CA certificates

```bash
docker build -t restaurant-agent .
docker run -p 8080:8080 --env-file .env restaurant-agent
```

The resulting image is minimal (roughly 15-20 MB).

### Toqan Platform

The project includes a detailed setup guide (`docs/toqan-setup-guide.md`) for deploying the agent on Toqan's web platform (work.toqan.ai). This involves:
1. Creating the agent with the system prompt
2. Configuring each Cube.js API tool as a custom integration
3. Setting up inventory via Google Sheets integration
4. Testing and publishing

---

## 10. Demo Scenarios

These demonstrate the agent's cross-domain reasoning:

| Scenario | Query | What It Does |
|----------|-------|--------------|
| Morning Briefing | "Give me my morning briefing" | Combines orders, top sellers, staff, labor costs, absences, and inventory alerts |
| Revenue Analysis | "How did we do this week? Break it down by channel." | Calls `get_revenue_by_channel` to compare dine-in, delivery, takeaway |
| Stock + Sales | "Top 5 sellers this week — do we have enough stock?" | Cross-references top sellers with inventory levels |
| Staff Check | "Who's working today? Are we staffed enough?" | Shows shifts by department with hours |
| Low Stock Alert | "What items are low? Place orders for everything." | Lists low stock, then places restock orders for each item |
| Menu Feasibility | "Can we serve 30 stamppot boerenkool for tonight?" | Checks each ingredient against stock for 30 servings |
