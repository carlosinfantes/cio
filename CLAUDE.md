# CTO Advisory Board - Developer Context

> This file provides context for AI assistants working on this codebase.

## Project Overview

CTO Advisory Board is a Go-based CLI tool that provides AI-powered advisory board consultations for technical decision-making. It features a dual-interface architecture supporting both CLI interactions and HTTP API for frontend integration.

## Architecture

### Dual Interface
- **CLI**: Interactive REPL and one-liner commands (`cto`, `cto ask`, `cto serve`)
- **API**: HTTP server with SSE streaming for React frontend integration

### Key Components

```
cmd/cto-advisory/main.go     - Entry point
internal/
├── api/                     - HTTP API server
│   ├── server.go           - Routes, handlers, session management
│   └── streaming.go        - SSE (Server-Sent Events) support
├── cli/
│   ├── commands/           - CLI commands (serve_cmd.go, etc.)
│   ├── output/             - Terminal formatting, mode selector
│   └── repl/               - Interactive REPL with facilitation
├── core/
│   ├── facilitation/       - Jordan facilitator state machine
│   │   ├── state.go        - Facilitation phases
│   │   ├── analyzer.go     - Message analysis
│   │   └── coordinator.go  - Auto-escalation logic
│   └── advisors/           - Persona definitions
├── plugins/                - Plugin system
│   ├── manifest.go         - Plugin schema
│   └── registry.go         - Plugin loading/management
└── storage/                - Data persistence
    ├── storage.go          - Interface abstraction
    └── file_storage.go     - File-based implementation
```

### Facilitation State Machine

Jordan (the facilitator) manages a state machine:
```
init → context_gathering → problem_articulation → discovery → ready_for_escalation → escalated
```

Auto-escalation triggers when:
- `ContextComplete = true`
- `ProblemArticulated = true`
- `DiscoveryComplete = true`

### Plugin System

Plugins define domain-specific advisory boards:
- Located in `plugins/` directory
- Each plugin has a `manifest.yaml`
- Loaded via `internal/plugins/registry.go`

Example plugin: `plugins/legal-advisory/manifest.yaml`

## Data Formats

### CRF (Context Reasoning Format)
Organizational context stored in `.cto-advisory/context/`:
- `organization.yaml` - Company info
- `teams.yaml` - Team structure
- `systems.yaml` - Tech stack
- `facts.yaml` - Constraints

### DRF (Decision Reasoning Format)
Decision records stored in `.cto-advisory/decisions/`

## Building & Running

```bash
# Build
make build

# Run CLI
./cto-advisory

# Run API server
./cto-advisory serve --port 8765

# Frontend development
cd frontend && npm install && npm run dev
```

## Key Files to Know

| File | Purpose |
|------|---------|
| `internal/api/server.go` | HTTP API routes and handlers |
| `internal/api/streaming.go` | SSE streaming implementation |
| `internal/core/facilitation/coordinator.go` | Facilitation logic |
| `internal/core/facilitation/state.go` | State machine definition |
| `internal/plugins/registry.go` | Plugin loading |
| `internal/storage/storage.go` | Storage interface |
| `frontend/src/hooks/useChat.ts` | React chat hook |
| `frontend/src/hooks/useStream.ts` | React SSE hook |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/session` | Create chat session |
| GET | `/api/v1/session/{id}` | Get session details |
| POST | `/api/v1/chat/{id}/message` | Send message |
| GET | `/api/v1/stream/{id}` | SSE streaming |
| GET | `/api/v1/context` | Get CRF entities |
| GET | `/api/v1/decisions` | List decisions |
| POST | `/api/v1/panel/ask` | Direct panel query |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_URL` | `http://localhost:8765` | API URL for frontend |

## Code Patterns

### Adding a New CLI Command
1. Create file in `internal/cli/commands/`
2. Register in `cmd/cto-advisory/main.go`

### Adding a New API Endpoint
1. Add handler in `internal/api/server.go`
2. Register route in `registerRoutes()`

### Creating a Plugin
1. Create directory in `plugins/`
2. Add `manifest.yaml` with facilitator, advisors, specialists
3. Load via plugin registry

## Testing

```bash
# Run tests
go test ./...

# Build check
go build ./...
```

## Dependencies

- Go 1.22+
- Cobra (CLI framework)
- Standard library for HTTP server
- React + TypeScript for frontend
- Vite for frontend build
- Tailwind CSS for styling
