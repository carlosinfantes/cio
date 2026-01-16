// Package types defines shared types for the CTO Advisory Board CLI.
package types

import "time"

// DiscoveryStatus represents the lifecycle of a discovery session.
type DiscoveryStatus string

const (
	DiscoveryStatusActive    DiscoveryStatus = "active"
	DiscoveryStatusConverted DiscoveryStatus = "converted"
	DiscoveryStatusAbandoned DiscoveryStatus = "abandoned"
)

// DiscoveryMessage represents a single turn in discovery conversation.
type DiscoveryMessage struct {
	Role      string    `yaml:"role"` // "facilitator" or "user"
	Content   string    `yaml:"content"`
	Timestamp time.Time `yaml:"timestamp"`
}

// DiscoverySession captures the full discovery conversation.
type DiscoverySession struct {
	ID             string             `yaml:"id"`
	Messages       []DiscoveryMessage `yaml:"messages"`
	Status         DiscoveryStatus    `yaml:"status"`
	CreatedAt      time.Time          `yaml:"created_at"`
	UpdatedAt      time.Time          `yaml:"updated_at"`
	GeneratedBrief *Brief             `yaml:"generated_brief,omitempty"`
}

// Brief is the structured template generated from discovery.
type Brief struct {
	ProblemStatement  string      `yaml:"problem_statement"`
	Context           string      `yaml:"context"`
	Constraints       []string    `yaml:"constraints"`
	Goals             []string    `yaml:"goals"`
	KeyQuestions      []string    `yaml:"key_questions"`
	SuggestedAdvisors []AdvisorID `yaml:"suggested_advisors"`
}

// NewDiscoverySession creates a new discovery session with initial facilitator greeting.
func NewDiscoverySession() *DiscoverySession {
	now := time.Now()
	return &DiscoverySession{
		ID:        generateDiscoveryID(now),
		Messages:  []DiscoveryMessage{},
		Status:    DiscoveryStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddMessage appends a message to the discovery session.
func (ds *DiscoverySession) AddMessage(role, content string) {
	ds.Messages = append(ds.Messages, DiscoveryMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	ds.UpdatedAt = time.Now()
}

// GetConversationText returns the full conversation as text for LLM context.
func (ds *DiscoverySession) GetConversationText() string {
	var text string
	for _, msg := range ds.Messages {
		if msg.Role == "facilitator" {
			text += "Facilitator: " + msg.Content + "\n\n"
		} else {
			text += "User: " + msg.Content + "\n\n"
		}
	}
	return text
}

// generateDiscoveryID creates a unique ID for a discovery session.
func generateDiscoveryID(t time.Time) string {
	return "disc-" + t.Format("2006-01-02-150405")
}
