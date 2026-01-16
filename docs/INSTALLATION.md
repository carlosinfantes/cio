# Installation Guide

## Requirements

- **Go 1.22+** (for building from source)
- **Node.js 18+** (for frontend development)
- **OpenRouter API key** or compatible LLM API

## Installation Methods

### Method 1: From Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/carlosinfantes/cto-advisory-board.git
cd cto-advisory-board

# Build the binary
make build

# Optional: Install globally
sudo mv cto-advisory /usr/local/bin/cto
```

### Method 2: Binary Releases

Download pre-built binaries from [GitHub Releases](https://github.com/carlosinfantes/cto-advisory-board/releases).

**macOS (Intel)**
```bash
curl -LO https://github.com/carlosinfantes/cto-advisory-board/releases/latest/download/cto-advisory-darwin-amd64
chmod +x cto-advisory-darwin-amd64
sudo mv cto-advisory-darwin-amd64 /usr/local/bin/cto
```

**macOS (Apple Silicon)**
```bash
curl -LO https://github.com/carlosinfantes/cto-advisory-board/releases/latest/download/cto-advisory-darwin-arm64
chmod +x cto-advisory-darwin-arm64
sudo mv cto-advisory-darwin-arm64 /usr/local/bin/cto
```

**Linux (x86_64)**
```bash
curl -LO https://github.com/carlosinfantes/cto-advisory-board/releases/latest/download/cto-advisory-linux-amd64
chmod +x cto-advisory-linux-amd64
sudo mv cto-advisory-linux-amd64 /usr/local/bin/cto
```

### Method 3: Go Install

```bash
go install github.com/carlosinfantes/cto-advisory-board/cmd/cto-advisory@latest
```

## Verify Installation

```bash
cto --version
# Output: cto-advisory version X.X.X
```

## Initial Setup

### 1. Run the Setup Wizard

```bash
cto init
```

This wizard will:
1. Configure your API key
2. Set up company context
3. Define team structure
4. Document tech stack
5. Record constraints and facts

### 2. Directory Structure Created

After initialization, you'll have:

```
.cto-advisory/
├── config.yaml              # API key and preferences
└── context/
    ├── organization.yaml    # Company info
    ├── teams.yaml           # Team structure
    ├── systems.yaml         # Tech stack
    └── facts.yaml           # Constraints
```

## Frontend Setup (Optional)

If you want to use the React frontend:

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The frontend will be available at `http://localhost:5173`.

To connect to the backend API, start the server:

```bash
cto serve --port 8765
```

## Docker Installation (Coming Soon)

```bash
docker pull carlosinfantes/cto-advisory-board
docker run -p 8765:8765 -v ~/.cto-advisory:/root/.cto-advisory carlosinfantes/cto-advisory-board
```

## Troubleshooting

### "command not found: cto"

Ensure the binary is in your PATH:
```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export PATH=$PATH:/usr/local/bin
```

### "permission denied"

Make sure the binary is executable:
```bash
chmod +x /usr/local/bin/cto
```

### Build errors

Ensure you have Go 1.22+:
```bash
go version
# Should output: go version go1.22.x or higher
```

### API connection errors

1. Verify your API key is set:
   ```bash
   cto config get api-key
   ```

2. Test the connection:
   ```bash
   cto ask "Hello" --verbose
   ```

## Updating

### From Source
```bash
cd cto-advisory-board
git pull
make build
sudo mv cto-advisory /usr/local/bin/cto
```

### Binary Releases
Download the latest release and replace the existing binary.

## Uninstalling

```bash
# Remove binary
sudo rm /usr/local/bin/cto

# Remove configuration (optional)
rm -rf ~/.cto-advisory

# Remove project-specific configuration (optional)
rm -rf .cto-advisory
```

## Next Steps

- [Configuration Guide](CONFIGURATION.md) - Customize your setup
- [Usage Guide](USAGE.md) - Learn the CLI commands
- [Architecture](ARCHITECTURE.md) - Understand the system
