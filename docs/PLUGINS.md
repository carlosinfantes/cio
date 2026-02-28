# Plugin Development Guide

## Overview

The plugin system allows you to create domain-specific advisory boards. Each plugin defines its own facilitator, advisors, specialists, and context entity types. Plugins can be installed from the central registry or created locally.

## Plugin Registry

CIO has a central plugin registry with 13 official plugins:

```bash
# Browse all plugins
cio plugin search

# Search by keyword
cio plugin search "finance"

# Install
cio plugin install startup-advisory

# Activate
cio plugin use startup-advisory
```

## Plugin Structure

```
my-domain/
├── manifest.yaml          # Required: plugin definition
├── settings.yaml          # Optional: plugin settings
├── personas/              # Optional: extended persona files
│   └── specialists/
├── cognitive/             # Optional: reasoning frameworks
│   └── frameworks/
└── prompts/               # Optional: custom prompts
```

## Manifest Schema

### Basic Example

```yaml
domain: my-domain
version: "1.0.0"
display_name: "My Domain Advisory Board"
description: "Advisory board for my specific domain"
author: "Your Name"
license: "MIT"
emoji: "🎯"

facilitator:
  id: "guide"
  name: "Guide Name"
  role: "Domain Intake Specialist"
  emoji: "💭"
  thinking_style: "Socratic questioning to draw out the full picture"

core_advisors:
  - id: "expert-1"
    name: "Expert One"
    role: "Senior Advisor"
    emoji: "🎯"
    thinking_style: "How does this fit the domain requirements?"
    priorities:
      - Quality
      - Efficiency

specialists:
  - id: "specialist-1"
    name: "Specialist One"
    role: "Specialist Role"
    emoji: "🔧"
    keywords:
      - keyword1
      - keyword2

decision_domains:
  - strategy
  - operations
```

### Complete Example

```yaml
domain: legal-advisory
version: "1.0.0"
display_name: "Legal Advisory Board"
description: "AI-powered legal advisory for business decisions"
author: "CIO Team"
emoji: "⚖️"

facilitator:
  id: "alex"
  name: "Alex Rivera"
  role: "Legal Intake Specialist"
  emoji: "💭"
  personality: |
    Professional, methodical, ensures all relevant facts are gathered.
    Asks about jurisdiction, parties involved, timeline, and documentation.

core_advisors:
  - id: "corporate-counsel"
    name: "Margaret Chen"
    role: "General Counsel"
    emoji: "⚖️"
    color: "blue"
    thinking_style: "What's the corporate governance angle?"
    expertise:
      - Corporate governance
      - M&A transactions

  - id: "litigation"
    name: "James Morrison"
    role: "Head of Litigation"
    emoji: "🛡️"
    color: "red"
    thinking_style: "What's our exposure and how do we protect against it?"

  - id: "contracts"
    name: "Sarah Kim"
    role: "Commercial Contracts Lead"
    emoji: "📝"
    color: "yellow"
    thinking_style: "What do the agreements actually say?"

  - id: "compliance"
    name: "Robert Williams"
    role: "Chief Compliance Officer"
    emoji: "📋"
    color: "cyan"
    thinking_style: "Are we meeting our regulatory obligations?"

specialists:
  - id: "ip-counsel"
    name: "Jennifer Wu"
    role: "IP Counsel"
    emoji: "💡"
    keywords: [patent, trademark, copyright, trade secret, licensing, IP]

  - id: "employment"
    name: "Michael Davis"
    role: "Employment Law Specialist"
    emoji: "👥"
    keywords: [employee, hiring, termination, discrimination, workplace, HR]

  - id: "privacy"
    name: "Emma Thompson"
    role: "Privacy & Data Protection"
    emoji: "🔒"
    keywords: [GDPR, CCPA, privacy, data protection, personal data, consent]

context_entities:
  - type: "client"
    description: "Client or company information"
    fields: [name, industry, jurisdiction, size]

  - type: "matter"
    description: "Legal matter or case"
    fields: [title, type, status, deadline]

escalation_criteria:
  context_required: [client, matter]
  questions_required:
    - "What is the primary legal issue?"
    - "What outcome are you seeking?"
```

## Manifest Fields Reference

### Root Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `domain` | string | Yes | Unique identifier |
| `version` | string | Yes | Semantic version (e.g., "1.0.0") |
| `display_name` | string | Yes | Human-readable name |
| `description` | string | No | Brief description |
| `author` | string | No | Plugin author |
| `license` | string | No | License type |
| `emoji` | string | No | Plugin emoji for display |
| `homepage` | string | No | Project URL |

### Facilitator

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `name` | string | Yes | Display name |
| `role` | string | Yes | Role description |
| `emoji` | string | No | Display emoji |
| `color` | string | No | Terminal color code |
| `thinking_style` | string | No | Approach to facilitation |
| `personality` | string | No | Personality traits |

### Advisors (Core & Specialists)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `name` | string | Yes | Display name |
| `role` | string | Yes | Role title |
| `emoji` | string | No | Display emoji |
| `color` | string | No | Terminal color |
| `thinking_style` | string | Yes | Characteristic question/approach |
| `personality` | string | No | Detailed personality |
| `expertise` | array | No | Areas of expertise |
| `priorities` | array | No | Key priorities |
| `keywords` | array | Specialists only | Auto-summon triggers |

## Creating a Plugin

### Option 1: From Template

```bash
cio plugin create my-domain
```

This scaffolds the full directory structure with template files.

### Option 2: Manual

```bash
mkdir -p ~/.cio/plugins/custom/my-domain
```

Create `manifest.yaml` with the schema above.

### Option 3: Install and Customize

```bash
# Install an existing plugin as a starting point
cio plugin install legal-advisory

# Copy to custom directory and modify
cp -r ~/.cio/plugins/installed/legal-advisory ~/.cio/plugins/custom/my-custom
```

## Loading Plugins

### CLI

```bash
# Activate an installed plugin
cio plugin use legal-advisory

# Use a specific plugin for one session
cio --plugin legal-advisory
```

### Programmatically

```go
import "github.com/carlosinfantes/cio/internal/plugins"

registry := plugins.GetRegistry()
err := registry.LoadPlugin("plugins/legal-advisory")
registry.SetActive("legal-advisory")
plugin := registry.GetActive()
```

## Plugin Locations

Plugins are loaded from these locations (in order):

1. `~/.cio/plugins/installed/` — Registry plugins
2. `~/.cio/plugins/custom/` — Custom plugins
3. `./plugins/` — Project-local plugins
4. Built-in plugins (cio)

## Best Practices

### 1. Clear Personas

Each advisor should have:
- Distinct personality and expertise
- Characteristic thinking style
- Clear role differentiation from other advisors

### 2. Relevant Keywords

For specialists, choose keywords that:
- Are specific to their expertise
- Users would naturally use in questions
- Don't overlap heavily with other specialists

### 3. Appropriate Escalation Criteria

Set criteria that ensure:
- Sufficient context before panel discussion
- Key questions are answered
- Advisors have what they need to give useful advice

### 4. Emoji Convention

- Use a single emoji per persona for consistent display
- Choose emojis that relate to the advisor's role
- The plugin-level emoji appears in `cio plugin search` results

## Publishing to the Registry

To submit your plugin to the official registry:

1. Package your plugin as a `.tar.gz`:
   ```bash
   cd my-domain && tar czf my-domain-1.0.0.tar.gz *
   ```

2. Submit a pull request to [cio-plugin-registry](https://github.com/carlosinfantes/cio-plugin-registry)

3. Include an entry for `index.json` with your plugin metadata

## Troubleshooting

### Plugin Not Loading

1. Check manifest syntax: `cat manifest.yaml`
2. Verify required fields (`domain`, `version`, `display_name`, `facilitator`, `core_advisors`)
3. Check for duplicate domain names across plugin directories

### Specialists Not Auto-Summoning

1. Verify keywords are lowercase
2. Ensure `auto_summon_specialists: true` in config
3. Test with explicit `--include specialist-id`

## Next Steps

- [Architecture](ARCHITECTURE.md) — System internals
- [Configuration](CONFIGURATION.md) — Customize settings
- [Usage Guide](USAGE.md) — CLI commands
