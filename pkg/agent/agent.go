package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"restaurant-agent/pkg/llm"
	"restaurant-agent/pkg/tools"
)

const sessionTTL = 1 * time.Hour

type Agent struct {
	client        *llm.Client
	tools         []anthropic.ToolUnionParam
	registry      *tools.Registry
	maxIterations int

	mu          sync.RWMutex
	sessions    map[string][]anthropic.MessageParam
	lastAccess  map[string]time.Time
	sessionLocks map[string]*sync.Mutex
}

func New(client *llm.Client, registry *tools.Registry, maxIterations int) *Agent {
	return &Agent{
		client:        client,
		tools:         llm.BuildToolDefinitions(),
		registry:      registry,
		maxIterations: maxIterations,
		sessions:      make(map[string][]anthropic.MessageParam),
		lastAccess:    make(map[string]time.Time),
		sessionLocks:  make(map[string]*sync.Mutex),
	}
}

// getSessionLock returns a per-session mutex, creating one if needed.
func (a *Agent) getSessionLock(sessionID string) *sync.Mutex {
	a.mu.Lock()
	defer a.mu.Unlock()
	mu, ok := a.sessionLocks[sessionID]
	if !ok {
		mu = &sync.Mutex{}
		a.sessionLocks[sessionID] = mu
	}
	return mu
}

// StartCleanup runs a background goroutine that removes sessions idle for longer
// than sessionTTL. Cancel the context to stop the cleanup loop.
func (a *Agent) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.cleanExpiredSessions()
			}
		}
	}()
}

func (a *Agent) cleanExpiredSessions() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	for id, lastAccess := range a.lastAccess {
		if now.Sub(lastAccess) > sessionTTL {
			delete(a.sessions, id)
			delete(a.lastAccess, id)
			delete(a.sessionLocks, id)
			log.Printf("[agent] cleaned up expired session: %s", id)
		}
	}
}

type ChatResponse struct {
	Response  string `json:"response"`
	SessionID string `json:"session_id"`
	ToolCalls int    `json:"tool_calls"`
}

func (a *Agent) Chat(ctx context.Context, sessionID string, userMessage string) (*ChatResponse, error) {
	// Serialize concurrent calls to the same session
	sessionMu := a.getSessionLock(sessionID)
	sessionMu.Lock()
	defer sessionMu.Unlock()

	// Get or create session history
	a.mu.Lock()
	history := a.sessions[sessionID]
	history = append(history, anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))
	a.lastAccess[sessionID] = time.Now()
	a.mu.Unlock()

	toolCallCount := 0

	for i := 0; i < a.maxIterations; i++ {
		resp, err := a.client.API.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     a.client.Model,
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{Text: llm.SystemPrompt},
			},
			Messages: history,
			Tools:    a.tools,
		})
		if err != nil {
			// Save history so the session isn't lost on transient errors
			a.mu.Lock()
			a.sessions[sessionID] = history
			a.mu.Unlock()
			return nil, fmt.Errorf("claude API error: %w", err)
		}

		// Append the assistant response to history
		history = append(history, resp.ToParam())

		// Check for tool use blocks and execute them
		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			switch block.Type {
			case "tool_use":
				toolUse := block.AsToolUse()
				toolCallCount++
				log.Printf("[agent] tool call: %s (id: %s)", toolUse.Name, toolUse.ID)

				result, err := a.registry.Execute(ctx, toolUse.Name, toolUse.Input)
				if err != nil {
					// Return error as tool result so Claude can handle gracefully
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						toolUse.ID, fmt.Sprintf("Error executing tool: %s", err.Error()), true,
					))
				} else {
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						toolUse.ID, result, false,
					))
				}
			}
		}

		// If no tool calls, we have the final response
		if len(toolResults) == 0 {
			// Extract text from response
			text := extractText(resp)

			// Save updated history
			a.mu.Lock()
			a.sessions[sessionID] = history
			a.mu.Unlock()

			return &ChatResponse{
				Response:  text,
				SessionID: sessionID,
				ToolCalls: toolCallCount,
			}, nil
		}

		// Feed tool results back to Claude
		history = append(history, anthropic.NewUserMessage(toolResults...))
	}

	// Save history even on max iterations so the session isn't lost
	a.mu.Lock()
	a.sessions[sessionID] = history
	a.mu.Unlock()
	return nil, fmt.Errorf("agent exceeded maximum iterations (%d)", a.maxIterations)
}

func extractText(msg *anthropic.Message) string {
	var text string
	for _, block := range msg.Content {
		if block.Type == "text" {
			if text != "" {
				text += "\n"
			}
			text += block.Text
		}
	}
	return text
}

// GetHistory returns the conversation history for a session (for debugging).
func (a *Agent) GetHistory(sessionID string) []map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	history := a.sessions[sessionID]
	result := make([]map[string]interface{}, len(history))
	for i, msg := range history {
		data, _ := json.Marshal(msg)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		result[i] = m
	}
	return result
}

// ClearSession removes a session's conversation history.
func (a *Agent) ClearSession(sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.sessions, sessionID)
	delete(a.lastAccess, sessionID)
	delete(a.sessionLocks, sessionID)
}
