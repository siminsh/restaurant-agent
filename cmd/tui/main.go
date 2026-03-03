package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"restaurant-agent/internal/store"
	"restaurant-agent/pkg/agent"
	"restaurant-agent/pkg/config"
	"restaurant-agent/pkg/llm"
	"restaurant-agent/pkg/tools"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	inventoryStore := store.New()

	registry := tools.NewRegistry()
	tools.RegisterInventoryTools(registry, inventoryStore)

	llmClient := llm.NewClient(cfg.AnthropicAPIKey, cfg.ClaudeModel)
	ag := agent.New(llmClient, registry, cfg.MaxAgentIterations)

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	ag.StartCleanup(cleanupCtx)

	sessionID := uuid.New().String()

	// Redirect log output to a temp file so agent logs don't corrupt the TUI
	logFile, err := os.CreateTemp("", "restaurant-tui-*.log")
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	m := newModel(ag, sessionID)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
