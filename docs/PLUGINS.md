# Plugin Development Guide

## Overview

The plugin system allows you to create domain-specific advisory boards. Each plugin defines its own facilitator, advisors, specialists, and context entity types.

## Plugin Structure

```
plugins/
└── your-domain/
    └── manifest.yaml
```

## Manifest Schema

### Basic Example

```yaml
# plugins/your-domain/manifest.yaml
domain: your-domain
version: "1.0.0"
display_name: "Your Domain Advisory Board"
description: "Advisory board for your specific domain"

facilitator:
  id: "guide"
  name: "Guide Name"
  role: "Domain Intake Specialist"
  personality: "Friendly, thorough, asks clarifying questions"

core_advisors:
  - id: "expert-1"
    name: "Expert One"
    role: "Senior Advisor"
    thinking_style: "How does this fit the domain requirements?"

specialists:
  - id: "specialist-1"
    name: "Specialist One"
    role: "Specialist Role"
    keywords:
      - keyword1
      - keyword2
```

### Complete Example: Legal Advisory

```yaml
# plugins/legal-advisory/manifest.yaml
domain: legal-advisory
version: "1.0.0"
display_name: "Legal Advisory Board"
description: "AI-powered legal advisory for business decisions"

facilitator:
  id: "alex"
  name: "Alex Rivera"
  role: "Legal Intake Specialist"
  personality: |
    Professional, methodical, ensures all relevant facts are gathered.
    Asks about jurisdiction, parties involved, timeline, and documentation.
  prompt_additions: |
    Always ask about:
    - Jurisdiction and applicable law
    - Parties involved and their relationships
    - Timeline and deadlines
    - Existing documentation or agreements

core_advisors:
  - id: "corporate-counsel"
    name: "Margaret Chen"
    role: "General Counsel"
    color: "blue"
    thinking_style: "What's the corporate governance angle?"
    personality: |
      Experienced corporate attorney with 20 years in M&A and governance.
      Focuses on fiduciary duties, board responsibilities, and shareholder interests.
    expertise:
      - Corporate governance
      - M&A transactions
      - Board advisory
      - Shareholder relations

  - id: "litigation"
    name: "James Morrison"
    role: "Head of Litigation"
    color: "red"
    thinking_style: "What's our exposure and how do we protect against it?"
    personality: |
      Former federal prosecutor, now defense-focused.
      Always considers worst-case scenarios and evidence preservation.
    expertise:
      - Civil litigation
      - Dispute resolution
      - Risk assessment
      - Evidence strategy

  - id: "contracts"
    name: "Sarah Kim"
    role: "Commercial Contracts Lead"
    color: "yellow"
    thinking_style: "What do the agreements actually say?"
    personality: |
      Detail-oriented contracts specialist.
      Reads every clause and considers enforcement scenarios.
    expertise:
      - Contract drafting
      - Commercial agreements
      - Terms negotiation
      - SLA management

  - id: "compliance"
    name: "Robert Williams"
    role: "Chief Compliance Officer"
    color: "cyan"
    thinking_style: "Are we meeting our regulatory obligations?"
    personality: |
      Regulatory expert across multiple industries.
      Proactive about compliance frameworks and audit readiness.
    expertise:
      - Regulatory compliance
      - Policy development
      - Audit preparation
      - Risk management

specialists:
  - id: "ip-counsel"
    name: "Jennifer Wu"
    role: "IP Counsel"
    color: "magenta"
    thinking_style: "How do we protect and leverage our intellectual property?"
    keywords:
      - patent
      - trademark
      - copyright
      - trade secret
      - licensing
      - IP
    expertise:
      - Patent strategy
      - Trademark registration
      - Copyright protection
      - Licensing agreements

  - id: "employment"
    name: "Michael Davis"
    role: "Employment Law Specialist"
    color: "magenta"
    thinking_style: "What are the employment law implications?"
    keywords:
      - employee
      - hiring
      - termination
      - discrimination
      - workplace
      - HR
      - benefits
    expertise:
      - Employment contracts
      - Workplace policies
      - Discrimination defense
      - Benefits compliance

  - id: "privacy"
    name: "Emma Thompson"
    role: "Privacy & Data Protection"
    color: "magenta"
    thinking_style: "How does this impact data privacy compliance?"
    keywords:
      - GDPR
      - CCPA
      - privacy
      - data protection
      - personal data
      - consent
    expertise:
      - Privacy regulations
      - Data protection
      - Consent management
      - Cross-border transfers

context_entities:
  - type: "client"
    description: "Client or company information"
    fields:
      - name
      - industry
      - jurisdiction
      - size

  - type: "matter"
    description: "Legal matter or case"
    fields:
      - title
      - type
      - status
      - deadline

  - type: "jurisdiction"
    description: "Applicable jurisdiction"
    fields:
      - country
      - state
      - governing_law

escalation_criteria:
  context_required:
    - client
    - matter
    - jurisdiction
  questions_required:
    - "What is the primary legal issue?"
    - "What outcome are you seeking?"
    - "What is your timeline?"
```

## Manifest Fields Reference

### Root Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `domain` | string | Yes | Unique identifier for the plugin |
| `version` | string | Yes | Semantic version (e.g., "1.0.0") |
| `display_name` | string | Yes | Human-readable name |
| `description` | string | No | Brief description |

### Facilitator

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `name` | string | Yes | Display name |
| `role` | string | Yes | Role description |
| `personality` | string | No | Personality traits |
| `prompt_additions` | string | No | Additional prompt instructions |

### Advisors (Core & Specialists)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `name` | string | Yes | Display name |
| `role` | string | Yes | Role title |
| `color` | string | No | Terminal color |
| `thinking_style` | string | Yes | Characteristic question |
| `personality` | string | No | Detailed personality |
| `expertise` | array | No | Areas of expertise |
| `keywords` | array | Specialists only | Auto-summon triggers |

### Context Entities

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Entity type name |
| `description` | string | No | Entity description |
| `fields` | array | No | Required fields |

### Escalation Criteria

| Field | Type | Description |
|-------|------|-------------|
| `context_required` | array | Entity types needed before escalation |
| `questions_required` | array | Questions that must be answered |

## Loading Plugins

### Programmatically

```go
import "github.com/carlosinfantes/cto-advisory-board/internal/plugins"

// Get the registry
registry := plugins.GetRegistry()

// Load a plugin
err := registry.LoadPlugin("plugins/legal-advisory")

// Set as active
registry.SetActive("legal-advisory")

// Get active plugin
plugin := registry.GetActive()
```

### CLI

```bash
# Use a specific plugin
cto --plugin legal-advisory

# Set default plugin
cto config set default-plugin legal-advisory
```

## Plugin Locations

Plugins are loaded from these locations (in order):

1. `./plugins/` - Project-local plugins
2. `~/.cto-advisory/plugins/` - User-global plugins
3. Built-in plugins (cto-advisory)

## Creating a New Plugin

### Step 1: Create Directory

```bash
mkdir -p plugins/my-domain
```

### Step 2: Create Manifest

```bash
touch plugins/my-domain/manifest.yaml
```

### Step 3: Define Content

```yaml
domain: my-domain
version: "1.0.0"
display_name: "My Domain Advisory"
description: "Custom advisory for my domain"

facilitator:
  id: "guide"
  name: "Domain Guide"
  role: "Intake Specialist"

core_advisors:
  - id: "advisor-1"
    name: "Advisor One"
    role: "Senior Expert"
    thinking_style: "What's the key consideration here?"
```

### Step 4: Test

```bash
cto --plugin my-domain
```

## Best Practices

### 1. Clear Personas

Each advisor should have:
- Distinct personality
- Specific expertise area
- Characteristic thinking style
- Clear role differentiation

### 2. Relevant Keywords

For specialists, choose keywords that:
- Are specific to their expertise
- Users would naturally use
- Don't overlap with other specialists

### 3. Thoughtful Context Entities

Define entities that:
- Capture essential domain information
- Have clear, minimal fields
- Support the advisory process

### 4. Appropriate Escalation Criteria

Set criteria that ensure:
- Sufficient context before panel discussion
- Key questions are answered
- Advisors have what they need

## Example Plugins

### Architecture Advisory

```yaml
domain: architecture-advisory
display_name: "Architecture Advisory Board"

facilitator:
  id: "intake"
  name: "Project Intake"
  role: "Architecture Intake Specialist"

core_advisors:
  - id: "systems"
    name: "Systems Architect"
    thinking_style: "What are the system boundaries and interfaces?"

  - id: "structural"
    name: "Structural Engineer"
    thinking_style: "What are the load-bearing requirements?"

  - id: "sustainability"
    name: "Sustainability Consultant"
    thinking_style: "What's the environmental impact?"
```

### HR Advisory

```yaml
domain: hr-advisory
display_name: "HR Advisory Board"

facilitator:
  id: "hr-intake"
  name: "HR Support"
  role: "HR Intake Specialist"

core_advisors:
  - id: "talent"
    name: "Talent Director"
    thinking_style: "How does this affect our people strategy?"

  - id: "compensation"
    name: "Compensation Manager"
    thinking_style: "What's the compensation and benefits angle?"

  - id: "culture"
    name: "Culture Lead"
    thinking_style: "How does this align with our values?"
```

## Troubleshooting

### Plugin Not Loading

1. Check manifest syntax:
   ```bash
   cat plugins/my-domain/manifest.yaml | yq .
   ```

2. Verify required fields are present

3. Check for duplicate domain names

### Specialists Not Auto-Summoning

1. Verify keywords are lowercase
2. Check keyword specificity
3. Test with explicit `--include`

### Context Entities Not Working

1. Verify entity type names match
2. Check field definitions
3. Review escalation criteria

## Next Steps

- [Architecture](ARCHITECTURE.md) - System internals
- [Configuration](CONFIGURATION.md) - Customize settings
- [Usage Guide](USAGE.md) - CLI commands
