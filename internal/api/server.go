// Package api implements the HTTP API server for the CTO Advisory Board.
// This enables building React frontends that communicate with the advisory engine.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/carlosinfantes/cto-advisory-board/internal/core/facilitation"
	"github.com/carlosinfantes/cto-advisory-board/internal/core/llm"
	"github.com/carlosinfantes/cto-advisory-board/internal/storage"
	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// Server is the HTTP API server.
type Server struct {
	addr     string
	store    storage.Storage
	client   *llm.Client
	sessions *SessionManager
	mux      *http.ServeMux
}

// SessionManager manages active chat sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// Session represents an active chat session.
type Session struct {
	ID          string                        `json:"id"`
	Domain      string                        `json:"domain"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
	State       *facilitation.FacilitationState `json:"-"`
	Coordinator *facilitation.Coordinator     `json:"-"`
	Messages    []ChatMessage                 `json:"messages"`
}

// ChatMessage represents a message in a session.
type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // user, jordan, panel, advisor_id
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewServer creates a new API server.
func NewServer(addr string, apiKey, model string) (*Server, error) {
	store, err := storage.GetDefaultStorage()
	if err != nil {
		return nil, fmt.Errorf("initializing storage: %w", err)
	}

	client, err := llm.NewClient(apiKey, model)
	if err != nil {
		return nil, fmt.Errorf("initializing LLM client: %w", err)
	}

	s := &Server{
		addr:   addr,
		store:  store,
		client: client,
		sessions: &SessionManager{
			sessions: make(map[string]*Session),
		},
		mux: http.NewServeMux(),
	}

	s.registerRoutes()
	return s, nil
}

func (s *Server) registerRoutes() {
	// CORS middleware wrapper
	cors := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			h(w, r)
		}
	}

	// Health check
	s.mux.HandleFunc("/api/health", cors(s.handleHealth))

	// Session management
	s.mux.HandleFunc("/api/v1/session", cors(s.handleSession))
	s.mux.HandleFunc("/api/v1/session/", cors(s.handleSessionByID))

	// Chat
	s.mux.HandleFunc("/api/v1/chat/", cors(s.handleChat))

	// Context (CRF)
	s.mux.HandleFunc("/api/v1/context", cors(s.handleContext))
	s.mux.HandleFunc("/api/v1/context/", cors(s.handleContextByID))

	// Decisions (DRF)
	s.mux.HandleFunc("/api/v1/decisions", cors(s.handleDecisions))
	s.mux.HandleFunc("/api/v1/decisions/", cors(s.handleDecisionByID))

	// Panel (direct access, skip Jordan)
	s.mux.HandleFunc("/api/v1/panel/ask", cors(s.handlePanelAsk))
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	fmt.Printf("Starting API server on %s\n", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

// === Health ===

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// === Sessions ===

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		s.createSession(w, r)
	case "GET":
		s.listSessions(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Domain = "cto-advisory"
	}

	// Load CRF context
	crfCtx, _ := s.store.LoadContext()

	// Create coordinator with callbacks
	callbacks := facilitation.CoordinatorCallbacks{
		// Callbacks will be handled differently in API mode
	}

	coordinator := facilitation.NewCoordinator(s.client, crfCtx, callbacks)

	session := &Session{
		ID:          uuid.New().String(),
		Domain:      req.Domain,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		State:       coordinator.GetState(),
		Coordinator: coordinator,
		Messages:    []ChatMessage{},
	}

	s.sessions.mu.Lock()
	s.sessions.sessions[session.ID] = session
	s.sessions.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"state":      session.State.Phase,
		"created_at": session.CreatedAt,
	})
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	s.sessions.mu.RLock()
	defer s.sessions.mu.RUnlock()

	sessions := make([]map[string]interface{}, 0, len(s.sessions.sessions))
	for _, sess := range s.sessions.sessions {
		sessions = append(sessions, map[string]interface{}{
			"id":           sess.ID,
			"domain":       sess.Domain,
			"created_at":   sess.CreatedAt,
			"updated_at":   sess.UpdatedAt,
			"message_count": len(sess.Messages),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
	})
}

func (s *Server) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/session/")
	if id == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	s.sessions.mu.RLock()
	session, ok := s.sessions.sessions[id]
	s.sessions.mu.RUnlock()

	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         session.ID,
			"domain":     session.Domain,
			"state":      session.State.Phase,
			"messages":   session.Messages,
			"created_at": session.CreatedAt,
			"updated_at": session.UpdatedAt,
		})
	case "DELETE":
		s.sessions.mu.Lock()
		delete(s.sessions.sessions, id)
		s.sessions.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// === Chat ===

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from path: /api/v1/chat/{session_id}/message
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/chat/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "message" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	sessionID := parts[0]

	s.sessions.mu.RLock()
	session, ok := s.sessions.sessions[sessionID]
	s.sessions.mu.RUnlock()

	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add user message
	userMsg := ChatMessage{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   req.Content,
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, userMsg)

	// Process with coordinator
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	result, err := session.Coordinator.ProcessMessage(ctx, req.Content)
	if err != nil {
		http.Error(w, fmt.Sprintf("Processing error: %v", err), http.StatusInternalServerError)
		return
	}

	// Add Jordan's response
	responseMsg := ChatMessage{
		ID:        uuid.New().String(),
		Role:      "jordan",
		Content:   result.Response,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"phase":           result.Phase,
			"ready_for_panel": result.ReadyForPanel,
			"escalated":       result.Escalated,
		},
	}
	session.Messages = append(session.Messages, responseMsg)
	session.UpdatedAt = time.Now()

	// Build response
	response := map[string]interface{}{
		"response": result.Response,
		"speaker":  "jordan",
		"state": map[string]interface{}{
			"phase":           result.Phase,
			"ready_for_panel": result.ReadyForPanel,
		},
	}

	if result.Escalated {
		response["escalated"] = true
		response["brief"] = result.Brief
	}

	if result.SuggestedMode != "" {
		response["suggested_mode"] = result.SuggestedMode
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// === Context ===

func (s *Server) handleContext(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ctx, err := s.store.LoadContext()
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading context: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if ctx == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"entities": []interface{}{},
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"entities": ctx.AllEntities(),
		})

	case "POST":
		var doc types.CRFDocument
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.store.SaveEntity(&doc); err != nil {
			http.Error(w, fmt.Sprintf("Saving entity: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      doc.Entity.ID,
			"created": true,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleContextByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/context/")
	if id == "" {
		http.Error(w, "Entity ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		entity, err := s.store.GetEntity(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading entity: %v", err), http.StatusInternalServerError)
			return
		}
		if entity == nil {
			http.Error(w, "Entity not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entity)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// === Decisions ===

func (s *Server) handleDecisions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		filter := &storage.DecisionFilter{}
		if status := r.URL.Query().Get("status"); status != "" {
			filter.Status = types.DRFStatus(status)
		}
		if tag := r.URL.Query().Get("tag"); tag != "" {
			filter.Tag = tag
		}
		if domain := r.URL.Query().Get("domain"); domain != "" {
			filter.Domain = domain
		}

		decisions, err := s.store.ListDecisions(filter)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading decisions: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"decisions": decisions,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDecisionByID(w http.ResponseWriter, r *http.Request) {
	// Handle /api/v1/decisions/{id} and /api/v1/decisions/{id}/status
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/decisions/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, "Decision ID required", http.StatusBadRequest)
		return
	}

	// Check for /status suffix
	if len(parts) > 1 && parts[1] == "status" {
		s.handleDecisionStatus(w, r, id)
		return
	}

	switch r.Method {
	case "GET":
		decision, err := s.store.GetDecision(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading decision: %v", err), http.StatusInternalServerError)
			return
		}
		if decision == nil {
			http.Error(w, "Decision not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(decision)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDecisionStatus(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateDecisionStatus(id, types.DRFStatus(req.Status)); err != nil {
		http.Error(w, fmt.Sprintf("Updating status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updated": true,
	})
}

// === Panel Direct Access ===

func (s *Server) handlePanelAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Question string   `json:"question"`
		Advisors []string `json:"advisors,omitempty"`
		Mode     string   `json:"mode,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Question == "" {
		http.Error(w, "Question is required", http.StatusBadRequest)
		return
	}

	// This would integrate with the modes package
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "panel_direct_not_implemented",
		"message": "Use the chat endpoint with a session for full functionality",
	})
}
