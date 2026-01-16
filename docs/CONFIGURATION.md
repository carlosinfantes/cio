# Configuration Guide

## Configuration File

All configuration is stored in `.cto-advisory/config.yaml`:

```yaml
# API Configuration
api_key: "sk-or-v1-..."           # OpenRouter API key
model: "anthropic/claude-3.5-sonnet"  # LLM model to use

# Default Behavior
default_mode: "panel"             # panel, socratic, advocate, framework
default_advisors:                 # Core advisors for panel discussions
  - cto
  - ciso
  - vp-eng
  - architect

# Facilitation Settings
auto_summon_specialists: true     # Auto-include specialists by keywords
max_advisors: 5                   # Soft cap, warns if exceeded

# Context Management
context_refresh_days: 30          # Prompt for review after N days
```

## Configuration Commands

### Set a Value

```bash
cto config set <key> <value>
```

**Examples:**
```bash
# Set API key
cto config set api-key YOUR_API_KEY

# Set default mode
cto config set default-mode socratic

# Set model
cto config set model anthropic/claude-3.5-sonnet
```

### Get a Value

```bash
cto config get <key>
```

**Example:**
```bash
cto config get default-mode
# Output: panel
```

### List All Settings

```bash
cto config list
```

## API Configuration

### OpenRouter (Recommended)

1. Sign up at [OpenRouter](https://openrouter.ai/)
2. Create an API key
3. Set the key:
   ```bash
   cto config set api-key sk-or-v1-your-key-here
   ```

### Anthropic Direct

```bash
cto config set api-key sk-ant-your-key-here
cto config set api-base https://api.anthropic.com
```

### Local LLM (Ollama)

```bash
cto config set api-base http://localhost:11434
cto config set model llama2
```

## Context Configuration

### Organization Context

File: `.cto-advisory/context/organization.yaml`

```yaml
crf_version: "0.1.0"
entity:
  id: "org-your-company"
  type: "organization"
  name: "Your Company"
  description: "Brief company description"
  attributes:
    org_type: "company"
    industry: "fintech"
    stage: "series-a"
    founded: 2022
    business_model: "B2B SaaS"
    size: "startup"
    compliance_frameworks:
      - soc2
      - pci-dss
```

### Team Context

File: `.cto-advisory/context/teams.yaml`

```yaml
crf_version: "0.1.0"
entity:
  id: "team-engineering"
  type: "organization"
  name: "Engineering Team"
  attributes:
    org_type: "team"
    headcount: 15
    skills:
      - backend
      - frontend
      - platform
    structure:
      backend: 6
      frontend: 4
      platform: 3
      mobile: 2
    unfilled_roles:
      - Staff Engineer
      - Security Lead
```

### Systems Context

File: `.cto-advisory/context/systems.yaml`

```yaml
crf_version: "0.1.0"
entity:
  id: "system-main"
  type: "system"
  name: "Main Platform"
  attributes:
    system_type: "platform"
    status: "production"
    hosting: "aws"
    primary_language: "typescript"
    languages:
      - typescript
      - python
      - go
    technology_stack:
      - ECS
      - Lambda
      - PostgreSQL
      - Redis
    deployment: "github-actions"
```

### Facts/Constraints

File: `.cto-advisory/context/facts.yaml`

```yaml
crf_version: "0.1.0"
entity:
  id: "fact-runway"
  type: "fact"
  name: "Financial Runway"
  description: "Current runway: 18 months"
  attributes:
    fact_type: "constraint"
    value: 18
    unit: "months"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CTO_API_KEY` | API key (overrides config) | - |
| `CTO_CONFIG_DIR` | Configuration directory | `.cto-advisory` |
| `VITE_API_URL` | Frontend API URL | `http://localhost:8765` |

**Example:**
```bash
export CTO_API_KEY="sk-or-v1-your-key"
cto ask "Should we adopt Kubernetes?"
```

## Advisor Configuration

### Default Advisors

```bash
cto config set default-advisors cto,ciso,architect
```

### Specialist Auto-Summoning

When `auto_summon_specialists: true`, these keywords trigger inclusion:

| Specialist | Keywords |
|------------|----------|
| CFO | budget, cost, ROI, pricing, revenue, burn, investment |
| Product | roadmap, feature, customers, market, MVP, product-market |
| DevOps | deploy, CI/CD, infrastructure, kubernetes, AWS, GCP |

Disable auto-summoning:
```bash
cto config set auto-summon-specialists false
```

## Mode Configuration

### Available Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `panel` | Full advisory board discussion | Complex decisions |
| `socratic` | Clarifying questions first | Vague problems |
| `advocate` | Challenge your decision | Validation |
| `framework` | Structured comparison | Option evaluation |

Set default:
```bash
cto config set default-mode socratic
```

Override per-query:
```bash
cto ask "AWS vs GCP" --mode framework
```

## Server Configuration

### API Server

```bash
# Start with default port
cto serve

# Custom port
cto serve --port 3000

# With CORS for development
cto serve --cors-origin http://localhost:5173
```

### Frontend Connection

Set the API URL for the frontend:

```bash
# In frontend/.env
VITE_API_URL=http://localhost:8765
```

Or at runtime:
```bash
VITE_API_URL=http://api.example.com npm run dev
```

## Plugin Configuration

### Load a Plugin

```bash
cto --plugin legal-advisory
```

### Set Default Plugin

```bash
cto config set default-plugin cto-advisory
```

### Plugin Directory

Plugins are loaded from:
1. `./plugins/` (project-local)
2. `~/.cto-advisory/plugins/` (user-global)

## Advanced Settings

### Timeout Configuration

```yaml
# In config.yaml
api_timeout: 120        # seconds
stream_timeout: 300     # seconds for streaming
```

### Logging

```bash
# Verbose output
cto ask "question" --verbose

# Debug mode
CTO_DEBUG=true cto ask "question"
```

### Cache Settings

```yaml
# In config.yaml
cache_decisions: true   # Cache recent decisions
cache_ttl: 3600        # Cache TTL in seconds
```

## Resetting Configuration

### Reset to Defaults

```bash
rm .cto-advisory/config.yaml
cto init
```

### Clear All Data

```bash
rm -rf .cto-advisory
cto init
```

## Next Steps

- [Usage Guide](USAGE.md) - Learn the CLI commands
- [Plugin Development](PLUGINS.md) - Create custom domains
- [Architecture](ARCHITECTURE.md) - System internals
