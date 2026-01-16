# Usage Guide

## CLI Interface

### Interactive Mode

Start an interactive session with Jordan (the facilitator):

```bash
cto
```

You'll see a mode selector:
```
╔══════════════════════════════════════════════════════════╗
║           CTO ADVISORY BOARD                              ║
╚══════════════════════════════════════════════════════════╝

Select mode:
  [1] Discover - Explore with Jordan
  [2] Decide   - Get panel advice
  [3] Challenge - Devil's advocate
  [4] Framework - Compare options
  [5] Context  - Review/update
  [6] History  - Browse decisions

Choice:
```

### Quick Questions (One-liners)

```bash
# Default panel mode
cto ask "Should we adopt Kubernetes?"

# With JSON output
cto ask "Should we adopt Kubernetes?" --json

# Specific mode
cto ask "AWS vs GCP" --mode framework

# Include specific advisors
cto ask "Security concerns with microservices" --advisors cto,ciso

# Force-include specialists
cto ask "Is this migration worth the cost?" --include cfo
```

## Interaction Modes

### Panel Mode (Default)

Full advisory board discussion with all perspectives:

```bash
cto ask "Should we migrate to microservices?"
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
cto ask "We need to scale" --mode socratic
```

**Flow:**
1. Jordan asks 3-5 clarifying questions
2. You provide answers
3. Panel discussion with enriched context

### Devil's Advocate Mode

Challenge an existing decision:

```bash
cto ask "We've decided to use MongoDB" --mode advocate
```

**Output includes:**
- Challenges from each advisor's perspective
- Blind spots identified
- Alternative approaches
- Risk assessment

### Framework Mode

Structured comparison of options:

```bash
cto ask "AWS vs GCP vs Azure" --mode framework
```

**Output includes:**
- Evaluation criteria
- Weighted scoring matrix
- Risk analysis per option
- Recommendation with confidence level

## Context Management

### View Context

```bash
# Summary view
cto context show

# Specific file
cto context show organization

# YAML format
cto context show --format yaml
```

### Edit Context

```bash
# Open in editor
cto context edit organization

# Check for staleness
cto context check

# AI-assisted update
cto context update
```

## Decision History

### List Decisions

```bash
# All decisions
cto history list

# Filter by status
cto history list --status approved

# Filter by tag
cto history list --tag security

# Search
cto history search "kubernetes"
```

### View Decision

```bash
cto history show dec-2024-01-12-001
```

### Update Decision

```bash
# Change status
cto history status dec-2024-01-12-001 approved

# Add tag
cto history tag dec-2024-01-12-001 infrastructure
```

## API Server

### Start the Server

```bash
# Default port (8765)
cto serve

# Custom port
cto serve --port 3000
```

### API Endpoints

#### Create Session

```bash
curl -X POST http://localhost:8765/api/v1/session
```

**Response:**
```json
{
  "session_id": "abc123",
  "phase": "init",
  "created_at": "2024-01-15T10:00:00Z"
}
```

#### Send Message

```bash
curl -X POST http://localhost:8765/api/v1/chat/abc123/message \
  -H "Content-Type: application/json" \
  -d '{"content": "Should we migrate to microservices?"}'
```

**Response:**
```json
{
  "response": "I'd like to understand your situation better...",
  "phase": "context_gathering",
  "ready_for_panel": false
}
```

#### Stream Responses (SSE)

```bash
curl http://localhost:8765/api/v1/stream/abc123
```

**Events:**
```
event: connected
data: {"session_id": "abc123"}

event: thinking
data: {"status": "processing"}

event: chunk
data: {"content": "I'd like to understand", "complete": false}

event: chunk
data: {"content": "I'd like to understand your situation", "complete": false}

event: complete
data: {"response": "...", "phase": "context_gathering", "ready_for_panel": false}
```

#### Get Context

```bash
curl http://localhost:8765/api/v1/context
```

#### List Decisions

```bash
curl http://localhost:8765/api/v1/decisions
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
cto ask "Should we adopt Kubernetes?"
```

### JSON

Machine-readable output:

```bash
cto ask "Should we adopt Kubernetes?" --json
```

**Output:**
```json
{
  "question": "Should we adopt Kubernetes?",
  "mode": "panel",
  "advisors": ["cto", "ciso", "vp-eng", "architect"],
  "responses": [
    {
      "advisor": "Victoria Chen",
      "role": "CTO",
      "content": "..."
    }
  ],
  "synthesis": "...",
  "decision_id": "dec-2024-01-15-001"
}
```

### Markdown

Clean markdown for documentation:

```bash
cto ask "Should we adopt Kubernetes?" --format markdown
```

## Working with Plugins

### List Available Plugins

```bash
cto plugins list
```

### Use a Specific Plugin

```bash
cto --plugin legal-advisory
```

### In a Session

```bash
cto
# Then select plugin when prompted
```

## Keyboard Shortcuts (Interactive Mode)

| Key | Action |
|-----|--------|
| `Ctrl+M` | Change mode |
| `Ctrl+C` | Exit |
| `Ctrl+D` | End input |
| `↑/↓` | Navigate history |

## Examples

### Quick Security Review

```bash
cto ask "Review the security implications of using JWT tokens" --advisors ciso,architect
```

### Cost Analysis

```bash
cto ask "What's the TCO of migrating to Kubernetes?" --include cfo
```

### Architecture Decision

```bash
cto ask "Monolith vs microservices for our stage" --mode framework
```

### Challenge a Decision

```bash
cto ask "We've decided to rewrite in Rust" --mode advocate
```

### Explore a Problem

```bash
cto ask "We're having scaling issues" --mode socratic
```

## Best Practices

1. **Provide Context**: The more context in your `.cto-advisory/context/` files, the more relevant the advice.

2. **Use Appropriate Modes**:
   - `panel` for complex decisions
   - `socratic` for unclear problems
   - `advocate` to validate decisions
   - `framework` for comparing options

3. **Track Decisions**: Use `--save` (default) to build decision history for future reference.

4. **Review Context Regularly**: Run `cto context check` monthly to keep context current.

5. **Include Relevant Specialists**: Use `--include` when you know budget, product, or infrastructure perspectives are needed.

## Next Steps

- [Configuration](CONFIGURATION.md) - Customize your setup
- [Plugin Development](PLUGINS.md) - Create custom domains
- [Architecture](ARCHITECTURE.md) - System internals
