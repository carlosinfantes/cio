# CTO Advisory Board - Architecture

## Overview

The CTO Advisory Board is a dual-interface system providing AI-powered advisory consultations:
- **CLI** - Command-line interface for human interaction
- **API** - HTTP endpoints for frontend/programmatic integration

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                       USER INTERFACES                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   CLI (Human)                         API (Frontend)                 │
│   ─────────────                       ──────────────                 │
│   $ cto                               POST /api/v1/session           │
│   $ cto ask "Q" --json                POST /api/v1/chat/{id}/message │
│   $ cto serve                         GET  /api/v1/stream/{id}       │
│                                                                      │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    FACILITATION LAYER                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   Jordan (Facilitator)                                               │
│   ────────────────────                                               │
│   internal/core/facilitation/                                        │
│                                                                      │
│   State Machine:                                                     │
│   ┌──────┐    ┌─────────┐    ┌─────────┐    ┌──────────┐    ┌─────┐│
│   │ init │ -> │ context │ -> │ problem │ -> │ discovery│ -> │panel││
│   └──────┘    └─────────┘    └─────────┘    └──────────┘    └─────┘│
│                                                                      │
│   Auto-escalation when: context + problem + discovery = complete    │
│                                                                      │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      PLUGIN SYSTEM                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   internal/plugins/                                                  │
│                                                                      │
│   plugins/                                                           │
│   ├── cto-advisory/           # Default: Tech decisions             │
│   │   └── manifest.yaml                                              │
│   │                                                                  │
│   ├── legal-advisory/         # Example: Legal decisions            │
│   │   └── manifest.yaml                                              │
│   │                                                                  │
│   └── [your-domain]/          # Custom domains                       │
│       └── manifest.yaml                                              │
│                                                                      │
│   Each plugin defines:                                               │
│   - Facilitator persona (Jordan equivalent)                          │
│   - Core advisors (always available)                                 │
│   - Specialists (auto-summoned by keywords)                          │
│   - Context entity types (domain-specific CRF)                       │
│                                                                      │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      STORAGE LAYER                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   internal/storage/                                                  │
│                                                                      │
│   Storage Interface                                                  │
│   ─────────────────                                                  │
│   - LoadContext() / SaveEntity()       # CRF operations             │
│   - SaveDecision() / GetDecision()     # DRF operations             │
│   - SaveDiscoverySession()             # Session persistence         │
│   - LoadConfig() / SaveConfig()        # Configuration               │
│                                                                      │
│   Implementations:                                                   │
│   ┌────────────────┐  ┌────────────────┐  ┌────────────────┐        │
│   │  FileStorage   │  │  SQLiteStorage │  │   APIStorage   │        │
│   │   (current)    │  │    (future)    │  │    (future)    │        │
│   └────────────────┘  └────────────────┘  └────────────────┘        │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Facilitation State Machine

**Location**: `internal/core/facilitation/`

Jordan's facilitation follows a state machine:

```go
type FacilitationPhase string

const (
    PhaseInit              FacilitationPhase = "init"
    PhaseContextGathering  FacilitationPhase = "context_gathering"
    PhaseProblemArticulation FacilitationPhase = "problem_articulation"
    PhaseDiscovery         FacilitationPhase = "discovery"
    PhaseReadyForEscalation FacilitationPhase = "ready_for_escalation"
    PhaseEscalated         FacilitationPhase = "escalated"
)
```

**Auto-escalation triggers when:**
- `ContextComplete = true`
- `ProblemArticulated = true`
- `DiscoveryComplete = true`

### 2. Storage Abstraction

**Location**: `internal/storage/`

```go
type Storage interface {
    // Context (CRF)
    LoadContext() (*types.CRFContext, error)
    SaveEntity(entity *types.CRFDocument) error

    // Decisions (DRF)
    SaveDecision(doc *types.DRFDocument) error
    GetDecision(id string) (*types.DRFDocument, error)

    // Sessions
    SaveDiscoverySession(session *types.DiscoverySession, name string) (string, error)
    LoadDiscoverySession(id string) (*types.DiscoverySession, error)
}
```

### 3. Plugin System

**Location**: `internal/plugins/`

Plugin manifest schema:

```yaml
domain: legal-advisory
version: "1.0.0"
display_name: "Legal Advisory Board"

facilitator:
  id: "alex"
  name: "Alex Rivera"
  role: "Legal Intake Specialist"

core_advisors:
  - id: "corporate-counsel"
    name: "Margaret Chen"
    role: "General Counsel"
    thinking_style: "What's the corporate governance angle?"

specialists:
  - id: "ip-counsel"
    name: "Jennifer Wu"
    keywords: [patent, trademark, copyright]

context_entities:
  - type: "client"
  - type: "matter"
  - type: "jurisdiction"
```

### 4. API Server

**Location**: `internal/api/`

Start with: `cto serve --port 8765`

**Endpoints:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/session` | Create chat session |
| GET | `/api/v1/session/{id}` | Get session details |
| POST | `/api/v1/chat/{id}/message` | Send message to Jordan |
| GET | `/api/v1/stream/{id}` | SSE streaming connection |
| GET | `/api/v1/context` | Get CRF entities |
| GET | `/api/v1/decisions` | List DRF decisions |
| POST | `/api/v1/panel/ask` | Direct panel query |

### 5. React Frontend

**Location**: `frontend/`

```bash
cd frontend
npm install
npm run dev
```

**Components:**
- `ChatPanel` - Main chat interface
- `ChatMessage` - Individual message display
- `ChatInput` - Message input with send
- `PhaseIndicator` - Facilitation progress

**Hooks:**
- `useChat` - Chat session management
- `useStream` - SSE streaming support

## Data Flow

```
User Input
    │
    ▼
┌───────────────────┐
│   Coordinator     │  ← Manages facilitation state
│   (facilitation)  │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│     Analyzer      │  ← Extracts info from messages
│   (facilitation)  │
└─────────┬─────────┘
          │
          ├──────────────────────────────┐
          │                              │
          ▼                              ▼
┌───────────────────┐        ┌───────────────────┐
│  Update State     │        │  Generate Response│
│  (FacilitationState)       │  (Jordan/LLM)     │
└─────────┬─────────┘        └─────────┬─────────┘
          │                              │
          ▼                              │
┌───────────────────┐                    │
│  Check Escalation │                    │
│  (auto-escalate?) │                    │
└─────────┬─────────┘                    │
          │                              │
          ├── Yes ───────────┐           │
          │                  │           │
          ▼                  ▼           │
┌───────────────────┐  ┌───────────────────┐
│  Continue Discovery│  │  Escalate to Panel│
│  (return response) │  │  (generate brief) │
└───────────────────┘  └───────────────────┘
```

## File Structure

```
ctoadvisoryboard/
├── cmd/cto-advisory/
│   └── main.go
├── internal/
│   ├── api/
│   │   ├── server.go        # HTTP API server
│   │   └── streaming.go     # SSE streaming
│   ├── cli/
│   │   ├── commands/
│   │   │   └── serve_cmd.go # cto serve command
│   │   ├── output/
│   │   │   └── mode_selector.go
│   │   └── repl/
│   │       └── enhanced.go  # Enhanced REPL
│   ├── core/
│   │   ├── facilitation/
│   │   │   ├── state.go     # State machine
│   │   │   ├── analyzer.go  # Message analysis
│   │   │   └── coordinator.go
│   │   └── advisors/        # Persona definitions
│   ├── plugins/
│   │   ├── manifest.go      # Plugin schema
│   │   └── registry.go      # Plugin loading
│   └── storage/
│       ├── storage.go       # Interface
│       └── file_storage.go  # Implementation
├── plugins/
│   └── legal-advisory/
│       └── manifest.yaml
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── api/
│   │   └── types/
│   └── package.json
└── docs/
    ├── ARCHITECTURE.md
    ├── INSTALLATION.md
    ├── CONFIGURATION.md
    ├── USAGE.md
    └── PLUGINS.md
```

## SSE Streaming Protocol

The API uses Server-Sent Events for real-time communication:

### Event Types

| Event | Data | Description |
|-------|------|-------------|
| `connected` | `{session_id}` | Connection established |
| `thinking` | `{status}` | Processing started |
| `chunk` | `{content, complete}` | Partial response |
| `complete` | `{response, phase, ready_for_panel}` | Response complete |
| `escalation` | `{brief}` | Panel escalation triggered |
| `heartbeat` | `{timestamp}` | Keep-alive (30s interval) |
| `error` | `{message}` | Error occurred |

### Frontend Integration

```typescript
// Connect to stream
const eventSource = new EventSource(`${API_BASE}/api/v1/stream/${sessionId}`);

eventSource.addEventListener('chunk', (event) => {
  const data = JSON.parse(event.data);
  setContent(data.content);
});

eventSource.addEventListener('complete', (event) => {
  const data = JSON.parse(event.data);
  // Handle completion
});
```

## Advisory Board Personas

### Core Advisors

| ID | Name | Role | Color | Focus |
|----|------|------|-------|-------|
| `cto` | Victoria Chen | Fractional CTO | Blue | Strategy, 10x outcomes |
| `ciso` | Marcus Webb | Former CISO | Red | Security, risk |
| `vp-eng` | Priya Sharma | VP Engineering | Yellow | Teams, delivery |
| `architect` | Erik Lindqvist | Principal Architect | Cyan | Trade-offs, design |

### Specialists (Auto-summoned)

| ID | Name | Keywords |
|----|------|----------|
| `cfo` | David Park | budget, cost, ROI, pricing |
| `product` | Sarah Mitchell | roadmap, feature, customers |
| `devops` | Alex Petrov | deploy, kubernetes, AWS |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_URL` | `http://localhost:8765` | API server URL for frontend |

## Future Roadmap

### Storage Implementations
- **SQLiteStorage** - Local database for improved querying
- **APIStorage** - Remote storage for team collaboration

### Plugin Enhancements
- Hot-reload plugins without restart
- Plugin marketplace
- Remote plugin loading

### Frontend Features
- Decision history browser
- Context editor
- Panel visualization
