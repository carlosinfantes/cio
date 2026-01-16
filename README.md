# CTO Advisory Board

> Your AI-powered executive committee for technical decisions

**Stop making critical decisions alone.** Get perspectives from a virtual CTO, CISO, VP Engineering, and Staff Architect — all debating your specific situation.

## Features

- **CLI-first** - One-liner commands for quick decisions, interactive sessions for deep exploration
- **API-ready** - HTTP endpoints for React/frontend integration
- **Jordan Facilitator** - AI facilitator that guides discovery before panel escalation
- **Plugin Architecture** - Extensible to any domain (legal, architecture, etc.)
- **Context-aware** - Uses CRF (Context Reasoning Format) for persistent organizational knowledge
- **Decision History** - DRF (Decision Reasoning Format) for decision documentation

## Quick Start

```bash
# Build
make build

# Initialize your project
./cto-advisory init

# Start interactive session with Jordan
./cto-advisory

# Ask a quick question
./cto-advisory ask "Should we adopt Kubernetes?" --json

# Start API server for frontend integration
./cto-advisory serve --port 8765
```

## Installation

### From Source

```bash
git clone https://github.com/carlosinfantes/cto-advisory-board.git
cd cto-advisory-board
make build
sudo mv cto-advisory /usr/local/bin/cto
```

### From Binary Releases

Download the binary for your platform from [Releases](https://github.com/carlosinfantes/cto-advisory-board/releases).

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                       USER INTERFACES                                │
├─────────────────────────────────────────────────────────────────────┤
│   CLI (Human)                         API (Frontend)                 │
│   $ cto                               POST /api/v1/session           │
│   $ cto ask "Q" --json                POST /api/v1/chat/{id}/message │
│   $ cto serve                         GET  /api/v1/stream/{id}       │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    FACILITATION LAYER                                │
│   Jordan (Facilitator) - Ensures context completeness               │
│   Auto-escalates to panel when ready                                │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      ADVISORY PANEL                                  │
│   Victoria Chen (CTO) | Marcus Webb (CISO)                          │
│   Priya Sharma (VP Eng) | Erik Lindqvist (Architect)                │
└─────────────────────────────────────────────────────────────────────┘
```

## Usage Modes

### Interactive Mode (Default)

```bash
$ cto

# Mode selector appears:
# [1] Discover - Explore with Jordan
# [2] Decide   - Get panel advice
# [3] Challenge - Devil's advocate
# [4] Framework - Compare options
# [5] Context  - Review/update
# [6] History  - Browse decisions
```

### One-liner Commands

```bash
# Quick question with JSON output
$ cto ask "Should we adopt Kubernetes?" --json

# Specific mode
$ cto ask "AWS vs GCP" --mode framework

# Export context
$ cto context show --format yaml

# List decisions
$ cto history list --status approved
```

### API Server

```bash
# Start server
$ cto serve --port 8765

# Create session
$ curl -X POST http://localhost:8765/api/v1/session

# Send message
$ curl -X POST http://localhost:8765/api/v1/chat/{session_id}/message \
    -H "Content-Type: application/json" \
    -d '{"content": "Should we migrate to microservices?"}'

# Stream responses (SSE)
$ curl http://localhost:8765/api/v1/stream/{session_id}
```

## The Advisory Board

### Core Advisors (Always Available)

| Advisor | Focus | Thinking Style |
|---------|-------|----------------|
| **Victoria Chen** (CTO) | Strategy, roadmap | "What's the 10x outcome?" |
| **Marcus Webb** (CISO) | Security, compliance | "What could go wrong?" |
| **Priya Sharma** (VP Eng) | Teams, execution | "Can we ship this?" |
| **Erik Lindqvist** (Architect) | System design | "Let me draw the trade-offs" |

### Specialists (Auto-summoned)

- **David Park** (CFO) — budget, costs, ROI
- **Sarah Mitchell** (Product) — customers, features
- **Alex Petrov** (DevOps) — infrastructure, deployment

## Facilitation Flow

Jordan, the facilitator, guides users through discovery before escalating to the panel:

```
User Input → Jordan asks clarifying questions → Context gathered
    ↓
Context complete? + Problem articulated? + Discovery done?
    ↓
Auto-escalate to Advisory Panel for decision
```

## Plugin System

Create custom advisory boards for any domain:

```yaml
# plugins/legal-advisory/manifest.yaml
domain: legal-advisory
display_name: "Legal Advisory Board"

facilitator:
  id: "alex"
  name: "Alex Rivera"
  role: "Legal Intake Specialist"

core_advisors:
  - id: "corporate-counsel"
    name: "Margaret Chen"
    role: "General Counsel"
```

Load plugins:
```bash
$ cto --plugin legal-advisory
```

## Configuration

Configuration stored in `.cto-advisory/config.yaml`:

```bash
# Set API key
cto config set api-key YOUR_OPENROUTER_API_KEY

# Set default mode
cto config set default-mode socratic

# List all settings
cto config list
```

## Project Context (CRF Format)

```
.cto-advisory/
├── config.yaml              # API key, preferences
├── context/
│   ├── organization.yaml    # Company stage, industry, compliance
│   ├── teams.yaml           # Team size, skills, structure
│   ├── systems.yaml         # Tech stack, hosting, languages
│   └── facts.yaml           # Constraints, runway, deadlines
└── decisions/               # Decision history (DRF format)
```

## Frontend Development

A React frontend is available for web-based interaction:

```bash
cd frontend
npm install
npm run dev
```

Connect to the API server at `http://localhost:8765` (configurable via `VITE_API_URL`).

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/session` | Create chat session |
| GET | `/api/v1/session/{id}` | Get session details |
| POST | `/api/v1/chat/{id}/message` | Send message to Jordan |
| GET | `/api/v1/stream/{id}` | SSE streaming connection |
| GET | `/api/v1/context` | Get CRF entities |
| GET | `/api/v1/decisions` | List DRF decisions |
| POST | `/api/v1/panel/ask` | Direct panel query |

## Requirements

- Go 1.22+ (for building from source)
- OpenRouter API key (or compatible LLM API)

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - System design and components
- [Installation](docs/INSTALLATION.md) - Detailed setup instructions
- [Configuration](docs/CONFIGURATION.md) - All configuration options
- [Usage Guide](docs/USAGE.md) - CLI and API examples
- [Plugin Development](docs/PLUGINS.md) - Creating custom domains

## License

MIT

## Built By

**Carlos Infantes** — [thewisecto.com](https://thewisecto.com)
