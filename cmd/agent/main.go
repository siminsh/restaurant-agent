package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"restaurant-agent/internal/store"
	"restaurant-agent/pkg/agent"
	"restaurant-agent/pkg/api"
	"restaurant-agent/pkg/config"
	"restaurant-agent/pkg/llm"
	"restaurant-agent/pkg/tools"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize components
	inventoryStore := store.New()
	log.Printf("inventory store initialized with %d items", len(inventoryStore.ListAll()))

	registry := tools.NewRegistry()
	tools.RegisterInventoryTools(registry, inventoryStore)
	log.Println("tool registry initialized with inventory tools")

	llmClient := llm.NewClient(cfg.AnthropicAPIKey, cfg.ClaudeModel)
	log.Printf("LLM client initialized (model: %s)", cfg.ClaudeModel)

	ag := agent.New(llmClient, registry, cfg.MaxAgentIterations)

	// Start background session cleanup
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	ag.StartCleanup(cleanupCtx)

	router := api.NewRouter(ag, inventoryStore)

	// Start HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("server shutdown failed: %v", err)
		}
	}()

	log.Printf("Restaurant operations agent running on http://localhost:%s", cfg.Port)
	log.Println("endpoints:")
	log.Println("  POST /api/v1/chat          - chat with the agent")
	log.Println("  GET  /api/v1/inventory      - list all inventory items")
	log.Println("  DELETE /api/v1/sessions/{id} - clear a chat session")
	log.Println("  GET  /health                - health check")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}

	log.Println("server stopped")
}
