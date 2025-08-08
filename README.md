# prom2grafana

A web application that intelligently converts raw Prometheus metrics into production-ready Grafana dashboards and alert rules using AI.

## Features

- 🚀 **Instant Conversion** - Paste Prometheus metrics, get a complete Grafana dashboard
- 🎯 **Smart Panel Generation** - AI creates appropriate visualizations for each metric type
- ⚡ **Alert Rules** - Automatically generates Prometheus alert rules based on metrics
- 🎨 **Clean UI** - Minimal, focused interface with real-time conversion
- 📋 **Export Ready** - Copy JSON directly or download dashboard files

## Demo

🚀 **[Try it live at prom2grafana.mrkaran.dev](https://prom2grafana.mrkaran.dev/)**

![prom2grafana demo](demo.gif)

## Installation

### Using Go

```bash
go install github.com/mr-karan/prom2grafana@latest
```

### Using Docker

```bash
docker run -p 8080:8080 -e OPENAI_API_KEY=your_key ghcr.io/mr-karan/prom2grafana
```

### From Source

```bash
git clone https://github.com/mr-karan/prom2grafana.git
cd prom2grafana
go mod download
go build -o prom2grafana
```

## Usage

1. Set your OpenRouter API key and configure models:
   ```bash
   export OPENAI_API_KEY=your_openrouter_api_key
   
   # Optional: Configure model fallback for better reliability
   export OPENAI_MODELS="google/gemini-2.5-flash,openai/gpt-4o,anthropic/claude-3-5-sonnet"
   ```

2. Run the server:
   ```bash
   ./prom2grafana
   ```

3. Open http://localhost:8080 in your browser

4. Paste your Prometheus metrics and click "Generate Dashboard"

5. Copy the generated Grafana dashboard JSON and import it into Grafana

### Model Fallback
The application automatically tries multiple models in order if configured. If the first model fails (rate limits, errors, etc.), it will automatically try the next model in the list, providing better reliability and success rates.

## Configuration

The application can be configured using environment variables:

### Basic Configuration
- `OPENAI_API_KEY` - Your OpenRouter API key (required)
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: info)

### Model Configuration
- `OPENAI_MODEL` - Single model to use (default: "google/gemini-2.5-flash")
- `OPENAI_MODELS` - Comma-separated list of models to try with fallback (overrides OPENAI_MODEL)
- `OPENAI_API_URL` - API URL (default: "https://openrouter.ai/api/v1")

### Model Fallback Examples
```bash
# Use single model (backward compatible)
export OPENAI_MODEL="google/gemini-2.5-flash"

# Use multiple models with fallback (recommended)
export OPENAI_MODELS="google/gemini-2.5-flash,openai/gpt-4o,anthropic/claude-3-5-sonnet"

# Cost-optimized fallback (cheapest first)
export OPENAI_MODELS="google/gemini-2.0-flash-thinking-exp,google/gemini-2.5-flash,openai/gpt-4o-mini"
```

## Example Input

```
# HELP node_cpu_seconds_total Seconds the CPUs spent in each mode.
# TYPE node_cpu_seconds_total counter
node_cpu_seconds_total{cpu="0",mode="idle"} 425836.82
node_cpu_seconds_total{cpu="0",mode="iowait"} 52.34

# HELP node_memory_MemAvailable_bytes Memory information field MemAvailable_bytes.
# TYPE node_memory_MemAvailable_bytes gauge
node_memory_MemAvailable_bytes 1.648762880e+10
```

## Development

```bash
# Install dependencies
go mod tidy

# Run with hot reload
air

# Run tests
go test ./...

# Build
just build

# Release
just release
```

## Tech Stack

- **Backend**: Go with net/http
- **Frontend**: Vanilla HTML/CSS/JS
- **AI**: Google Gemini 2.0 Flash via OpenRouter
- **Deployment**: Single binary with embedded assets

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details

## Author

[mr-karan](https://github.com/mr-karan)