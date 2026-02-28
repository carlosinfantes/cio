# CIO - Chief Intelligence Officer

> AI-powered advisory boards for intelligent decision-making

**Stop making critical decisions alone.** Assemble a virtual advisory board of domain experts who debate your specific situation — from C-suite strategy to legal counsel, startup fundraising to personal finance.

## Features

- **CLI-first** — One-liner commands for quick decisions, interactive REPL for deep exploration
- **API-ready** — HTTP server with SSE streaming for React/frontend integration
- **Jordan Facilitator** — AI facilitator that guides discovery before panel escalation
- **Plugin Registry** — 13 official advisory boards, install any domain in seconds
- **Context-aware** — CRF (Context Reasoning Format) for persistent organizational knowledge
- **Decision History** — DRF (Decision Reasoning Format) for decision documentation

## Quick Start

```bash
# Install from source
git clone https://github.com/carlosinfantes/cio.git
cd cio && make build

# Initialize your project
./cio init

# Ask a question
./cio ask "Should we adopt Kubernetes?"

# Start interactive session with Jordan
./cio

# Start API server for frontend
./cio serve --port 8765
```

## Plugin Registry

Browse and install domain-specific advisory boards:

```bash
# Search available plugins
$ cio plugin search

  🚀 startup-advisory ★ (1.0.0)
    Startup Advisory Board
    AI-powered advisory for founders and early-stage companies
    ★ 421  ↓ 8.2K

  📣 marketing-advisory ★ (1.0.0)
    Marketing & Brand Advisory Board
    AI-powered CMO-level counsel for marketing strategy
    ★ 534  ↓ 14.0K

# Install a plugin
$ cio plugin install startup-advisory

# Activate it
$ cio plugin use startup-advisory
```

### Official Plugins

| Plugin | Domain | Description |
|--------|--------|-------------|
| 💭 `cio` | Technology | CTO-level executive committee for tech decisions |
| ⚖️ `legal-advisory` | Legal | Corporate counsel, contracts, compliance |
| 📊 `financial-advisory` | Business | CFO-level financial strategy and operations |
| 🚀 `startup-advisory` | Business | Founders, VCs, operators, go-to-market |
| 📣 `marketing-advisory` | Business | CMO-level marketing and brand strategy |
| 💡 `product-advisory` | Business | CPO-level product strategy and growth |
| 🤗 `people-advisory` | Business | CHRO-level people strategy and culture |
| 🧠 `data-ai-advisory` | Technology | Data strategy, ML, and AI governance |
| 🔍 `security-advisory` | Technology | CISO-level cybersecurity strategy |
| 🧭 `career-advisory` | Personal | Career transitions, negotiation, growth |
| 🧭 `personal-finance` | Personal | Investing, budgeting, retirement planning |
| ✨ `creative-advisory` | Personal | Creators, writers, and indie builders |
| 🌱 `wellness-advisory` | Personal | Health, fitness, nutrition, mental health |

## Usage Modes

### Interactive Mode

```bash
$ cio

# Mode selector:
# [1] Discover  — Explore with Jordan
# [2] Decide    — Get panel advice
# [3] Challenge — Devil's advocate
# [4] Framework — Compare options
# [5] Context   — Review/update
# [6] History   — Browse decisions
```

### One-liner Commands

```bash
# Quick question with JSON output
$ cio ask "Should we adopt Kubernetes?" --json

# Specific mode
$ cio ask "AWS vs GCP" --mode framework

# Devil's advocate
$ cio ask "We've decided to use MongoDB" --mode advocate

# With specific advisors
$ cio ask "Platform review" --advisors cto,architect
```

### API Server

```bash
# Start server
$ cio serve --port 8765

# Create session
$ curl -X POST http://localhost:8765/api/v1/session

# Send message
$ curl -X POST http://localhost:8765/api/v1/chat/{id}/message \
    -H "Content-Type: application/json" \
    -d '{"content": "Should we migrate to microservices?"}'

# Stream responses (SSE)
$ curl http://localhost:8765/api/v1/stream/{id}
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    USER INTERFACES                      │
│   CLI (cio, cio ask)          API (cio serve)           │
│   Interactive REPL            POST /api/v1/chat/        │
│                               GET  /api/v1/stream/      │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                 FACILITATION LAYER                       │
│   Jordan — guides discovery, gathers context            │
│   Auto-escalates when context + problem + discovery     │
│   are complete                                          │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                   ADVISORY PANEL                        │
│   Core advisors debate from different perspectives      │
│   Specialists auto-summoned based on keywords           │
│   Plugin system swaps entire advisory boards            │
└─────────────────────────────────────────────────────────┘
```

## Facilitation Flow

Jordan guides users through discovery before escalating to the panel:

```
User Input → Jordan asks clarifying questions → Context gathered
    ↓
Context complete? + Problem articulated? + Discovery done?
    ↓
Auto-escalate to Advisory Panel for decision
```

## Plugin Development

Create custom advisory boards for any domain:

```yaml
# my-plugin/manifest.yaml
domain: my-domain
display_name: "My Advisory Board"
version: "1.0.0"

facilitator:
  id: "facilitator"
  name: "Jordan"
  role: "Discovery Coach"
  emoji: "💭"

core_advisors:
  - id: "expert-1"
    name: "Expert One"
    role: "Domain Expert"
    emoji: "🎯"
    thinking_style: "Strategic, long-term perspective"
```

```bash
# Create from template
$ cio plugin create my-domain

# Or load directly
$ cio --plugin my-domain
```

## Configuration

```bash
# Set API key
cio config set api-key YOUR_API_KEY

# Set default mode
cio config set default-mode socratic

# List all settings
cio config list
```

Configuration stored in `.cio/config.yaml`:

```
.cio/
├── config.yaml              # API key, preferences
├── context/
│   ├── organization.yaml    # Company stage, industry
│   ├── teams.yaml           # Team size, skills
│   ├── systems.yaml         # Tech stack, hosting
│   └── facts.yaml           # Constraints, deadlines
├── decisions/               # Decision history (DRF)
└── plugins/
    ├── installed/           # Registry plugins
    └── custom/              # Your custom plugins
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/session` | Create chat session |
| GET | `/api/v1/session/{id}` | Get session details |
| POST | `/api/v1/chat/{id}/message` | Send message to Jordan |
| POST | `/api/v1/chat/stream/{id}` | Send message with SSE response |
| GET | `/api/v1/stream/{id}` | SSE streaming connection |
| GET | `/api/v1/context` | Get CRF entities |
| GET | `/api/v1/decisions` | List DRF decisions |
| POST | `/api/v1/panel/ask` | Direct panel query (skip Jordan) |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CIO_API_KEY` | — | OpenRouter API key |
| `CIO_CONFIG_DIR` | `.cio` | Config directory |
| `CIO_DEBUG` | `false` | Enable debug mode |
| `VITE_API_URL` | `http://localhost:8765` | API URL for frontend |

## Requirements

- Go 1.22+ (building from source)
- OpenRouter API key (or compatible LLM API)

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — System design and components
- [Installation](docs/INSTALLATION.md) — Detailed setup instructions
- [Configuration](docs/CONFIGURATION.md) — All configuration options
- [Usage Guide](docs/USAGE.md) — CLI and API examples
- [Plugin Development](docs/PLUGINS.md) — Creating custom domains

## License

[MIT](LICENSE)

## Built By

**Carlos Infantes**
