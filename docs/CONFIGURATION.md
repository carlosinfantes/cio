# Configuration Guide

## Configuration File

All configuration is stored in `.cio/config.yaml`:

```yaml
# API Configuration
api_key: "sk-or-v1-..."              # OpenRouter API key
model: "anthropic/claude-3.5-sonnet"  # LLM model to use

# Default Behavior
default_mode: "panel"                # panel, socratic, advocate, framework
default_advisors:
  - cto
  - ciso
  - vp-eng
  - architect

# Facilitation Settings
auto_summon_specialists: true        # Auto-include specialists by keywords
max_advisors: 5                      # Soft cap, warns if exceeded

# Context Management
context_refresh_days: 30             # Prompt for review after N days

# Plugin Settings
active_domain: ""                    # Active plugin domain
registry_url: ""                     # Custom registry URL (default: GitHub)
```

## Configuration Commands

### Set a Value

```bash
cio config set <key> <value>
```

**Examples:**
```bash
cio config set api-key YOUR_API_KEY
cio config set default-mode socratic
cio config set model anthropic/claude-3.5-sonnet
```

### Get a Value

```bash
cio config get <key>
```

### List All Settings

```bash
cio config list
```

## API Configuration

### OpenRouter (Recommended)

1. Sign up at [OpenRouter](https://openrouter.ai/)
2. Create an API key
3. Set the key:
   ```bash
   cio config set api-key sk-or-v1-your-key-here
   ```

### Anthropic Direct

```bash
cio config set api-key sk-ant-your-key-here
cio config set api-base https://api.anthropic.com
```

### Local LLM (Ollama)

```bash
cio config set api-base http://localhost:11434
cio config set model llama2
```

## Context Configuration

### Organization Context

File: `.cio/context/organization.yaml`

```yaml
crf_version: "0.1.0"
entity:
  id: "org-your-company"
  type: "organization"
  name: "Your Company"
  attributes:
    org_type: "company"
    industry: "fintech"
    stage: "series-a"
    business_model: "B2B SaaS"
    size: "startup"
    compliance_frameworks:
      - soc2
      - pci-dss
```

### Team Context

File: `.cio/context/teams.yaml`

```yaml
crf_version: "0.1.0"
entity:
  id: "team-engineering"
  type: "organization"
  name: "Engineering Team"
  attributes:
    org_type: "team"
    headcount: 15
    skills: [backend, frontend, platform]
    unfilled_roles:
      - Staff Engineer
      - Security Lead
```

### Systems Context

File: `.cio/context/systems.yaml`

```yaml
crf_version: "0.1.0"
entity:
  id: "system-main"
  type: "system"
  name: "Main Platform"
  attributes:
    hosting: "aws"
    primary_language: "typescript"
    technology_stack: [ECS, Lambda, PostgreSQL, Redis]
    deployment: "github-actions"
```

### Facts/Constraints

File: `.cio/context/facts.yaml`

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
| `CIO_API_KEY` | API key (overrides config) | — |
| `CIO_CONFIG_DIR` | Configuration directory | `.cio` |
| `CIO_DEBUG` | Enable debug mode | `false` |
| `VITE_API_URL` | Frontend API URL | `http://localhost:8765` |

```bash
export CIO_API_KEY="sk-or-v1-your-key"
cio ask "Should we adopt Kubernetes?"
```

## Advisor Configuration

### Default Advisors

```bash
cio config set default-advisors cto,ciso,architect
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
cio config set auto-summon-specialists false
```

## Mode Configuration

| Mode | Description | Use Case |
|------|-------------|----------|
| `panel` | Full advisory board discussion | Complex decisions |
| `socratic` | Clarifying questions first | Vague problems |
| `advocate` | Challenge your decision | Validation |
| `framework` | Structured comparison | Option evaluation |

```bash
# Set default
cio config set default-mode socratic

# Override per-query
cio ask "AWS vs GCP" --mode framework
```

## Server Configuration

```bash
# Default port (8765)
cio serve

# Custom port
cio serve --port 3000
```

### Frontend Connection

```bash
# In frontend/.env
VITE_API_URL=http://localhost:8765
```

## Plugin Configuration

```bash
# Install from registry
cio plugin install legal-advisory

# Activate
cio plugin use legal-advisory

# Set via flag
cio --plugin legal-advisory
```

Plugin directories:
1. `~/.cio/plugins/installed/` — Registry plugins
2. `~/.cio/plugins/custom/` — Custom plugins
3. `./plugins/` — Project-local plugins

## Resetting Configuration

```bash
# Reset config only
rm .cio/config.yaml
cio init

# Clear all data
rm -rf .cio
cio init
```

## Next Steps

- [Usage Guide](USAGE.md) — Learn the CLI commands
- [Plugin Development](PLUGINS.md) — Create custom domains
- [Architecture](ARCHITECTURE.md) — System internals
