// Package api implements Server-Sent Events (SSE) streaming for real-time responses.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// StreamEvent represents an SSE event.
type StreamEvent struct {
	ID    string      `json:"id,omitempty"`
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// StreamConnection represents an active SSE connection.
type StreamConnection struct {
	SessionID string
	Channel   chan StreamEvent
	Done      chan struct{}
}

// StreamManager manages SSE connections.
type StreamManager struct {
	mu          sync.RWMutex
	connections map[string]*StreamConnection
}

// NewStreamManager creates a new stream manager.
func NewStreamManager() *StreamManager {
	return &StreamManager{
		connections: make(map[string]*StreamConnection),
	}
}

// Connect creates a new streaming connection for a session.
func (sm *StreamManager) Connect(sessionID string) *StreamConnection {
	conn := &StreamConnection{
		SessionID: sessionID,
		Channel:   make(chan StreamEvent, 100),
		Done:      make(chan struct{}),
	}

	sm.mu.Lock()
	sm.connections[sessionID] = conn
	sm.mu.Unlock()

	return conn
}

// Disconnect removes a streaming connection.
func (sm *StreamManager) Disconnect(sessionID string) {
	sm.mu.Lock()
	if conn, ok := sm.connections[sessionID]; ok {
		close(conn.Done)
		close(conn.Channel)
		delete(sm.connections, sessionID)
	}
	sm.mu.Unlock()
}

// Send sends an event to a session's stream.
func (sm *StreamManager) Send(sessionID string, event StreamEvent) bool {
	sm.mu.RLock()
	conn, ok := sm.connections[sessionID]
	sm.mu.RUnlock()

	if !ok {
		return false
	}

	select {
	case conn.Channel <- event:
		return true
	case <-conn.Done:
		return false
	default:
		// Channel full, drop event
		return false
	}
}

// Broadcast sends an event to all connections.
func (sm *StreamManager) Broadcast(event StreamEvent) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, conn := range sm.connections {
		select {
		case conn.Channel <- event:
		case <-conn.Done:
		default:
		}
	}
}

// Global stream manager
var streamManager = NewStreamManager()

// registerStreamRoutes registers the SSE streaming routes.
func (s *Server) registerStreamRoutes() {
	s.mux.HandleFunc("/api/v1/stream/", s.handleStream)
}

// handleStream handles SSE streaming connections.
func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path: /api/v1/stream/{session_id}
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/v1/stream/")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Check session exists
	s.sessions.mu.RLock()
	_, exists := s.sessions.sessions[sessionID]
	s.sessions.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create connection
	conn := streamManager.Connect(sessionID)
	defer streamManager.Disconnect(sessionID)

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial connection event
	writeSSEEvent(w, StreamEvent{
		Event: "connected",
		Data:  map[string]string{"session_id": sessionID},
	})
	flusher.Flush()

	// Heartbeat ticker
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Context for cancellation
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.Done:
			return
		case event := <-conn.Channel:
			writeSSEEvent(w, event)
			flusher.Flush()
		case <-heartbeat.C:
			writeSSEEvent(w, StreamEvent{
				Event: "heartbeat",
				Data:  map[string]int64{"timestamp": time.Now().Unix()},
			})
			flusher.Flush()
		}
	}
}

// writeSSEEvent writes a Server-Sent Event to the response.
func writeSSEEvent(w http.ResponseWriter, event StreamEvent) {
	if event.ID != "" {
		fmt.Fprintf(w, "id: %s\n", event.ID)
	}
	if event.Event != "" {
		fmt.Fprintf(w, "event: %s\n", event.Event)
	}

	data, err := json.Marshal(event.Data)
	if err != nil {
		data = []byte("{}")
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
}

// StreamingChat processes a chat message with streaming response.
func (s *Server) streamingChat(ctx context.Context, sessionID, content string) error {
	s.sessions.mu.RLock()
	session, ok := s.sessions.sessions[sessionID]
	s.sessions.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}

	// Send thinking event
	streamManager.Send(sessionID, StreamEvent{
		Event: "thinking",
		Data:  map[string]string{"status": "processing"},
	})

	// Process message
	result, err := session.Coordinator.ProcessMessage(ctx, content)
	if err != nil {
		streamManager.Send(sessionID, StreamEvent{
			Event: "error",
			Data:  map[string]string{"message": err.Error()},
		})
		return err
	}

	// Stream response chunks (simulate streaming for now)
	words := strings.Fields(result.Response)
	var accumulated strings.Builder

	for i, word := range words {
		accumulated.WriteString(word)
		if i < len(words)-1 {
			accumulated.WriteString(" ")
		}

		streamManager.Send(sessionID, StreamEvent{
			Event: "chunk",
			Data: map[string]interface{}{
				"content":  accumulated.String(),
				"complete": i == len(words)-1,
			},
		})

		// Small delay between chunks for visual effect
		time.Sleep(20 * time.Millisecond)
	}

	// Send completion event
	streamManager.Send(sessionID, StreamEvent{
		Event: "complete",
		Data: map[string]interface{}{
			"response":        result.Response,
			"phase":           result.Phase,
			"ready_for_panel": result.ReadyForPanel,
			"escalated":       result.Escalated,
		},
	})

	if result.Escalated && result.Brief != nil {
		streamManager.Send(sessionID, StreamEvent{
			Event: "escalation",
			Data:  result.Brief,
		})
	}

	return nil
}
