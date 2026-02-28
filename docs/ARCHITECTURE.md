# CIO - Architecture

## Overview

CIO is a dual-interface system providing AI-powered advisory consultations:
- **CLI** — Command-line interface for human interaction
- **API** — HTTP endpoints with SSE streaming for frontend integration

## System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    USER INTERFACES                      │
├─────────────────────────────────────────────────────────┤
│   CLI (Human)                    API (Frontend)         │
│   $ cio                         POST /api/v1/session    │
│   $ cio ask "Q" --json          POST /api/v1/chat/      │
│   $ cio serve                   GET  /api/v1/stream/    │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                 FACILITATION LAYER                       │
├─────────────────────────────────────────────────────────┤
│   Jordan (Facilitator)                                  │
│   internal/core/facilitation/                           │
│                                                         │
│   State Machine:                                        │
│   init → context → problem → discovery → panel          │
│                                                         │
│   Auto-escalation when:                                 │
│   context + problem + discovery = complete              │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                   PLUGIN SYSTEM                         │
├─────────────────────────────────────────────────────────┤
│   internal/plugins/                                     │
│   internal/plugins/remote/  (registry client)           │
│                                                         │
│   Plugin sources:                                       │
│   ├── Registry    (cio plugin install <domain>)         │
│   ├── Local       (~/.cio/plugins/installed/)           │
│   └── Custom      (~/.cio/plugins/custom/)              │
│                                                         │
│   Each plugin defines:                                  │
│   - Facilitator persona                                 │
│   - Core advisors (always available)                    │
│   - Specialists (auto-summoned by keywords)             │
│   - Context entity types (domain-specific CRF)          │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                    STORAGE LAYER                        │
├─────────────────────────────────────────────────────────┤
│   internal/storage/                                     │
│                                                         │
│   Storage Interface                                     │
│   - LoadContext() / SaveEntity()       # CRF            │
│   - SaveDecision() / GetDecision()     # DRF            │
│   - SaveDiscoverySession()             # Sessions       │
│   - LoadConfig() / SaveConfig()        # Configuration  │
│                                                         │
│   Implementation: FileStorage (file-based, goroutine-   │
│   safe singleton via sync.Once)                         │
└─────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Facilitation State Machine

**Location**: `internal/core/facilitation/`

```go
type FacilitationPhase string

const (
    PhaseInit                FacilitationPhase = "init"
    PhaseContextGathering    FacilitationPhase = "context_gathering"
    PhaseProblemArticulation FacilitationPhase = "problem_articulation"
    PhaseDiscovery           FacilitationPhase = "discovery"
    PhaseReadyForEscalation  FacilitationPhase = "ready_for_escalation"
    PhaseEscalated           FacilitationPhase = "escalated"
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
    LoadContext() (*types.CRFContext, error)
    SaveEntity(entity *types.CRFDocument) error
    SaveDecision(doc *types.DRFDocument) error
    GetDecision(id string) (*types.DRFDocument, error)
    SaveDiscoverySession(session *types.DiscoverySession, name string) (string, error)
    LoadDiscoverySession(id string) (*types.DiscoverySession, error)
}
```

### 3. Plugin System

**Location**: `internal/plugins/`, `internal/plugins/remote/`

Plugin manifest schema:

```yaml
domain: legal-advisory
version: "1.0.0"
display_name: "Legal Advisory Board"
emoji: "⚖️"

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
```

Plugin management:
```bash
cio plugin search              # Browse registry (stars, downloads)
cio plugin install <domain>    # Install from registry
cio plugin use <domain>        # Activate a plugin
cio plugin list                # List installed plugins
cio plugin create <domain>     # Scaffold a custom plugin
cio plugin update              # Update to latest versions
```

### 4. API Server

**Location**: `internal/api/`

Start with: `cio serve --port 8765`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/session` | Create chat session |
| GET | `/api/v1/session/{id}` | Get session details |
| POST | `/api/v1/chat/{id}/message` | Send message to Jordan |
| POST | `/api/v1/chat/stream/{id}` | Send message with SSE response |
| GET | `/api/v1/stream/{id}` | SSE streaming connection |
| GET | `/api/v1/context` | Get CRF entities |
| GET | `/api/v1/decisions` | List DRF decisions |
| POST | `/api/v1/panel/ask` | Direct panel query |

### 5. React Frontend

**Location**: `frontend/`

```bash
cd frontend && npm install && npm run dev
```

**Components:** `ChatPanel`, `ChatMessage`, `ChatInput`, `PhaseIndicator`
**Hooks:** `useChat` (session management), `useStream` (SSE streaming)

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
          ├───────────────────────┐
          ▼                       ▼
┌───────────────────┐   ┌───────────────────┐
│  Update State     │   │  Generate Response│
│  (FacilState)     │   │  (Jordan/LLM)     │
└─────────┬─────────┘   └─────────┬─────────┘
          │                       │
          ▼                       │
┌───────────────────┐             │
│  Check Escalation │             │
└─────────┬─────────┘             │
          │                       │
          ├── Yes ──┐             │
          ▼         ▼             │
┌──────────────┐ ┌────────────────┐
│  Continue    │ │  Escalate to   │
│  Discovery   │ │  Panel         │
└──────────────┘ └────────────────┘
```

## File Structure

```
cio/
├── cmd/cio/main.go                  # Entry point, version injection
├── internal/
│   ├── api/
│   │   ├── server.go                # HTTP API server, routes
│   │   └── streaming.go             # SSE streaming
│   ├── cli/
│   │   ├── commands/
│   │   │   ├── root.go              # Root command, loadPlugin()
│   │   │   ├── ask_cmd.go           # cio ask
│   │   │   ├── plugin_cmd.go        # cio plugin (search/install/etc)
│   │   │   ├── serve_cmd.go         # cio serve
│   │   │   ├── init.go              # cio init wizard
│   │   │   ├── config_cmd.go        # cio config
│   │   │   ├── context_cmd.go       # cio context
│   │   │   ├── history_cmd.go       # cio history
│   │   │   └── session.go           # Session management
│   │   ├── output/                  # Terminal formatting
│   │   ├── repl/                    # Interactive REPL
│   │   └── wizard/                  # Setup wizard
│   ├── config/                      # Configuration loading
│   ├── core/
│   │   ├── advisors/                # Persona definitions
│   │   ├── context/                 # CRF loading/saving
│   │   ├── decisions/               # DRF management
│   │   ├── discovery/               # Discovery sessions
│   │   ├── facilitation/            # State machine
│   │   ├── llm/                     # LLM client & prompts
│   │   └── modes/                   # Panel, socratic, etc.
│   ├── plugins/
│   │   ├── loader.go                # Plugin loading
│   │   ├── manifest.go              # YAML schema
│   │   ├── registry.go              # Plugin registry
│   │   └── remote/                  # Registry client, downloader
│   ├── storage/                     # File-based storage
│   └── types/                       # Shared types
├── plugins/cio/                     # Built-in CIO plugin
├── plugin-templates/                # Scaffolding templates
├── frontend/                        # React + TypeScript frontend
└── docs/                            # Documentation
```

## SSE Streaming Protocol

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
