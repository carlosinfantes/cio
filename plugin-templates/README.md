# Plugin Templates

This directory contains templates used when creating new plugins with `cto plugin create`.

## Structure

```
plugin-templates/
└── default/
    ├── manifest.yaml.tmpl      # Main plugin configuration
    ├── settings.yaml.tmpl      # Domain-specific settings
    ├── personas/
    │   └── facilitator.yaml.tmpl
    ├── cognitive/
    │   └── discovery.yaml.tmpl
    └── prompts/
        ├── greeting.md
        └── synthesis.md
```

## Template Variables

Templates use Go template syntax with the following variables:

| Variable | Description |
|----------|-------------|
| `{{.Domain}}` | Plugin domain identifier (e.g., "my-advisory") |
| `{{.DisplayName}}` | Human-readable name |
| `{{.Author}}` | Plugin author name |

## Creating Custom Templates

To create a new template set:

1. Create a new directory under `plugin-templates/`
2. Copy the `default/` structure
3. Modify templates as needed
4. Use `--template <name>` flag with `cto plugin create`

## Plugin Package Structure

When a user runs `cto plugin create`, the template is expanded to create:

```
.cto-advisory/plugins/custom/<domain>/
├── manifest.yaml          # From manifest.yaml.tmpl
├── settings.yaml          # From settings.yaml.tmpl
├── personas/
│   ├── facilitator.yaml   # From personas/facilitator.yaml.tmpl
│   └── specialists/       # Empty, user adds as needed
├── cognitive/
│   ├── discovery.yaml     # From cognitive/discovery.yaml.tmpl
│   └── frameworks/        # Empty, user adds as needed
└── prompts/
    ├── greeting.md        # Copied directly
    └── synthesis.md       # Copied directly
```
