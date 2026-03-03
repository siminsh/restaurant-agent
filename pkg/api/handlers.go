package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"restaurant-agent/internal/store"
	"restaurant-agent/pkg/agent"
)

type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func handleChat(ag *agent.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Limit request body to 1MB to prevent OOM
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Message == "" {
			writeError(w, http.StatusBadRequest, "message is required")
			return
		}

		if req.SessionID == "" {
			req.SessionID = uuid.New().String()
		}

		resp, err := ag.Chat(r.Context(), req.SessionID, req.Message)
		if err != nil {
			log.Printf("[api] chat error (session=%s): %v", req.SessionID, err)
			writeError(w, http.StatusInternalServerError, "an internal error occurred while processing your request")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handleListInventory(s *store.MemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items := s.ListAll()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": items,
			"count": len(items),
		})
	}
}

func handleClearSession(ag *agent.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionID")
		ag.ClearSession(sessionID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
