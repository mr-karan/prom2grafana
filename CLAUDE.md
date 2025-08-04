# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Tasks

### Build and Run
```bash
# Run the application locally
just run

# Build binary
just build

# Run with hot reload during development
just dev

# Run tests
just test

# Format and lint code
just fmt
just lint
```

### Release Process
```bash
# Create a git tag to trigger release
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# Test release locally
just release

# The GitHub Actions workflow will automatically:
# - Build binaries for Linux amd64
# - Create Docker images
# - Publish to GitHub Releases
```

## Code Architecture

This is a single-file Go web application (`main.go`) that converts Prometheus metrics to Grafana dashboards using AI.

### Key Components

1. **Embedded Assets**: Uses Go's `embed` directive to bundle:
   - `index.html` - Frontend interface
   - `static/style.css` - Styling
   - `grafana_dashboard_prompt.md` - AI prompt template

2. **HTTP Server**: Simple net/http server with:
   - `/` - Serves the web interface
   - `/convert` - API endpoint for metric conversion
   - `/static/` - Serves CSS files

3. **AI Integration**: 
   - Uses OpenRouter API (configurable via env vars)
   - Default model: Google Gemini 2.0 Flash
   - Structured output using JSON schema for consistent responses

4. **Configuration**:
   - Environment-based using koanf
   - Required: `OPENAI_API_KEY`
   - Optional: `PORT`, `OPENAI_API_URL`, `OPENAI_MODEL`

### Important Notes

- The application name changed from `grafana-dashboard-gen` to `prom2grafana`
- Uses structured logging with slog
- CORS is enabled for local development
- Version information is injected during build by GoReleaser
- All dependencies are minimal and focused on core functionality