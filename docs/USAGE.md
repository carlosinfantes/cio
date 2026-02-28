# Usage Guide

## CLI Interface

### Interactive Mode

Start an interactive session with Jordan (the facilitator):

```bash
cio
```

You'll see a mode selector:
```
╔══════════════════════════════════════════════════════════╗
║           CIO ADVISORY BOARD                            ║
╚══════════════════════════════════════════════════════════╝

Select mode:
  [1] Discover  — Explore with Jordan
  [2] Decide    — Get panel advice
  [3] Challenge — Devil's advocate
  [4] Framework — Compare options
  [5] Context   — Review/update
  [6] History   — Browse decisions

Choice:
```

### Quick Questions (One-liners)

```bash
# Default panel mode
cio ask "Should we adopt Kubernetes?"

# With JSON output
cio ask "Should we adopt Kubernetes?" --json

# Specific mode
cio ask "AWS vs GCP" --mode framework

# Include specific advisors
cio ask "Security concerns with microservices" --advisors cto,ciso

# Force-include specialists
cio ask "Is this migration worth the cost?" --include cfo
```

## Interaction Modes

### Panel Mode (Default)

Full advisory board discussion with all perspectives:

```bash
cio ask "Should we migrate to microservices?"
```

**Output includes:**
- Individual advisor perspectives
- Areas of consensus
- Points of tension
- Synthesized recommendation
- Key questions to consider

### Socratic Mode

Jordan asks clarifying questions before consulting the panel:

```bash
cio ask "We need to scale" --mode socratic
```

**Flow:**
1. Jordan asks 3-5 clarifying questions
2. You provide answers
3. Panel discussion with enriched context

### Devil's Advocate Mode

Challenge an existing decision:

```bash
cio ask "We've decided to use MongoDB" --mode advocate
```

**Output includes:**
- Challenges from each advisor's perspective
- Blind spots identified
- Alternative approaches
- Risk assessment

### Framework Mode

Structured comparison of options:

```bash
cio ask "AWS vs GCP vs Azure" --mode framework
```

**Output includes:**
- Evaluation criteria
- Weighted scoring matrix
- Risk analysis per option
- Recommendation with confidence level

## Plugin Management

### Browse the Registry

```bash
cio plugin search
```

Shows all available plugins with stars, downloads, and featured highlights.

### Search by Keyword

```bash
cio plugin search "finance"
```

### Install a Plugin

```bash
cio plugin install startup-advisory
```

### Activate a Plugin

```bash
cio plugin use startup-advisory
```

### List Installed Plugins

```bash
cio plugin list
```

### Get Plugin Details

```bash
cio plugin info startup-advisory
```

### Update Plugins

```bash
# Update all
cio plugin update

# Update specific
cio plugin update startup-advisory
```

### Create a Custom Plugin

```bash
cio plugin create my-domain
```

## Context Management

### View Context

```bash
# Summary view
cio context show

# Specific file
cio context show organization

# YAML format
cio context show --format yaml
```

### Edit Context

```bash
# Open in editor
cio context edit organization

# Check for staleness
cio context check

# AI-assisted update
cio context update
```

## Decision History

### List Decisions

```bash
# All decisions
cio history list

# Filter by status
cio history list --status approved

# Filter by tag
cio history list --tag security

# Search
cio history search "kubernetes"
```

### View Decision

```bash
cio history show dec-2024-01-12-001
```

### Update Decision

```bash
# Change status
cio history status dec-2024-01-12-001 approved

# Add tag
cio history tag dec-2024-01-12-001 infrastructure
```

## API Server

### Start the Server

```bash
# Default port (8765)
cio serve

# Custom port
cio serve --port 3000
```

### API Examples

#### Create Session

```bash
curl -X POST http://localhost:8765/api/v1/session
```

```json
{
  "session_id": "abc123",
  "phase": "init",
  "created_at": "2026-02-28T10:00:00Z"
}
```

#### Send Message

```bash
curl -X POST http://localhost:8765/api/v1/chat/abc123/message \
  -H "Content-Type: application/json" \
  -d '{"content": "Should we migrate to microservices?"}'
```

#### Stream Responses (SSE)

```bash
curl http://localhost:8765/api/v1/stream/abc123
```

```
event: connected
data: {"session_id": "abc123"}

event: chunk
data: {"content": "I'd like to understand your situation", "complete": false}

event: complete
data: {"response": "...", "phase": "context_gathering", "ready_for_panel": false}
```

#### Direct Panel Query

```bash
curl -X POST http://localhost:8765/api/v1/panel/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Should we adopt Kubernetes?",
    "mode": "panel",
    "advisors": ["cto", "ciso", "architect"]
  }'
```

## Output Formats

### Terminal (Default)

Rich formatted output with colored boxes:
```bash
cio ask "Should we adopt Kubernetes?"
```

### JSON

Machine-readable output:
```bash
cio ask "Should we adopt Kubernetes?" --json
```

### Markdown

Clean markdown for documentation:
```bash
cio ask "Should we adopt Kubernetes?" --format markdown
```

## Keyboard Shortcuts (Interactive Mode)

| Key | Action |
|-----|--------|
| `Ctrl+M` | Change mode |
| `Ctrl+C` | Exit |
| `Ctrl+D` | End input |
| `Up/Down` | Navigate history |

## Examples

### Quick Security Review

```bash
cio ask "Review security implications of using JWT tokens" --advisors ciso,architect
```

### Cost Analysis

```bash
cio ask "What's the TCO of migrating to Kubernetes?" --include cfo
```

### Architecture Decision

```bash
cio ask "Monolith vs microservices for our stage" --mode framework
```

### Challenge a Decision

```bash
cio ask "We've decided to rewrite in Rust" --mode advocate
```

### Explore a Problem

```bash
cio ask "We're having scaling issues" --mode socratic
```

## Best Practices

1. **Provide context** — The more context in `.cio/context/`, the more relevant the advice
2. **Use appropriate modes** — `panel` for decisions, `socratic` for unclear problems, `advocate` to validate, `framework` to compare
3. **Track decisions** — Decisions are saved by default for future reference
4. **Review context regularly** — Run `cio context check` monthly to keep context current
5. **Include specialists** — Use `--include` when budget, product, or infrastructure perspectives are needed

## Next Steps

- [Configuration](CONFIGURATION.md) — Customize your setup
- [Plugin Development](PLUGINS.md) — Create custom domains
- [Architecture](ARCHITECTURE.md) — System internals
