# prom2grafana

A web application that intelligently converts raw Prometheus metrics into production-ready Grafana dashboards and alert rules using AI.

## Features

- 🚀 **Instant Conversion** - Paste Prometheus metrics, get a complete Grafana dashboard
- 🎯 **Smart Panel Generation** - AI creates appropriate visualizations for each metric type
- ⚡ **Alert Rules** - Automatically generates Prometheus alert rules based on metrics
- 🎨 **Clean UI** - Minimal, focused interface with real-time conversion
- 📋 **Export Ready** - Copy JSON directly or download dashboard files

## Demo

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

1. Set your OpenRouter API key:
   ```bash
   export OPENAI_API_KEY=your_openrouter_api_key
   ```

2. Run the server:
   ```bash
   ./prom2grafana
   ```

3. Open http://localhost:8080 in your browser

4. Paste your Prometheus metrics and click "Generate Dashboard"

5. Copy the generated Grafana dashboard JSON and import it into Grafana

## Configuration

The application can be configured using environment variables:

- `OPENAI_API_KEY` - Your OpenRouter API key (required)
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: info)

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