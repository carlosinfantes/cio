# Installation Guide

## Requirements

- **Go 1.22+** (for building from source)
- **Node.js 18+** (for frontend development, optional)
- **OpenRouter API key** or compatible LLM API

## Installation Methods

### Method 1: From Source (Recommended)

```bash
git clone https://github.com/carlosinfantes/cio.git
cd cio
make build
sudo mv cio /usr/local/bin/cio
```

### Method 2: Binary Releases

Download pre-built binaries from [GitHub Releases](https://github.com/carlosinfantes/cio/releases).

**macOS (Apple Silicon)**
```bash
curl -LO https://github.com/carlosinfantes/cio/releases/latest/download/cio-darwin-arm64
chmod +x cio-darwin-arm64
sudo mv cio-darwin-arm64 /usr/local/bin/cio
```

**macOS (Intel)**
```bash
curl -LO https://github.com/carlosinfantes/cio/releases/latest/download/cio-darwin-amd64
chmod +x cio-darwin-amd64
sudo mv cio-darwin-amd64 /usr/local/bin/cio
```

**Linux (x86_64)**
```bash
curl -LO https://github.com/carlosinfantes/cio/releases/latest/download/cio-linux-amd64
chmod +x cio-linux-amd64
sudo mv cio-linux-amd64 /usr/local/bin/cio
```

### Method 3: Go Install

```bash
go install github.com/carlosinfantes/cio/cmd/cio@latest
```

## Verify Installation

```bash
cio version
# Output: cio version 1.0.0
```

## Initial Setup

### 1. Run the Setup Wizard

```bash
cio init
```

The wizard will:
1. Configure your API key
2. Set up company context
3. Define team structure
4. Document tech stack
5. Record constraints and facts

### 2. Directory Structure Created

```
.cio/
├── config.yaml              # API key and preferences
├── context/
│   ├── organization.yaml    # Company info
│   ├── teams.yaml           # Team structure
│   ├── systems.yaml         # Tech stack
│   └── facts.yaml           # Constraints
├── decisions/               # Decision history
└── plugins/
    ├── installed/           # Registry plugins
    └── custom/              # Your custom plugins
```

### 3. Install a Plugin

```bash
# Browse available plugins
cio plugin search

# Install one
cio plugin install startup-advisory

# Activate it
cio plugin use startup-advisory
```

## Frontend Setup (Optional)

```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at `http://localhost:5173`.

Start the backend API:
```bash
cio serve --port 8765
```

## Troubleshooting

### "command not found: cio"

Ensure the binary is in your PATH:
```bash
export PATH=$PATH:/usr/local/bin
```

### "permission denied"

```bash
chmod +x /usr/local/bin/cio
```

### Build errors

Ensure you have Go 1.22+:
```bash
go version
```

### API connection errors

1. Verify your API key:
   ```bash
   cio config get api-key
   ```

2. Test the connection:
   ```bash
   cio ask "Hello"
   ```

## Updating

### From Source
```bash
cd cio
git pull
make build
sudo mv cio /usr/local/bin/cio
```

### Update Plugins
```bash
cio plugin update
```

## Uninstalling

```bash
sudo rm /usr/local/bin/cio
rm -rf ~/.cio      # User configuration (optional)
rm -rf .cio        # Project configuration (optional)
```

## Next Steps

- [Configuration Guide](CONFIGURATION.md) — Customize your setup
- [Usage Guide](USAGE.md) — Learn the CLI commands
- [Plugin Development](PLUGINS.md) — Create custom domains
