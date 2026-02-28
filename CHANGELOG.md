# Changelog

## [1.0.0] - 2026-02-28

### Added
- Dual-interface architecture: CLI (interactive REPL + one-liner) and HTTP API with SSE streaming
- Jordan facilitator with state machine (init -> context_gathering -> problem_articulation -> discovery -> escalation)
- Advisory board modes: panel, socratic, advocate, framework
- Plugin system with YAML manifests for domain-specific advisory boards
- Remote plugin registry with search, install, update, and uninstall
- Plugin stars and download counts in search results
- CRF (Context Reasoning Format) for organizational context
- DRF (Decision Reasoning Format) for decision records
- Interactive setup wizard (`cio init`)
- Decision history with search and status tracking
- Context staleness detection and refresh prompts
- Auto-summoning of specialist advisors based on keywords
- React + TypeScript frontend with SSE streaming support
- 13 official advisory board plugins: CIO, legal, financial, startup, data-ai, security, people, product, marketing, personal-finance, wellness, career, creative
- Cross-platform builds (macOS, Linux, Windows)
