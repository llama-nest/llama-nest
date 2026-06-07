# Local Development Guide

This guide will help you set up your local development environment for `llama-nest`.

## Prerequisites

Before you start, ensure you have the following installed:

### Required

- **Go 1.22 or higher** — The project is built with Go 1.22. [Install Go](https://go.dev/dl/)
  - Verify installation: `go version`

- **Ollama** — llama-nest is a wrapper around Ollama. [Install Ollama](https://ollama.ai/)
  - Verify installation: `ollama --version`
  - Ensure Ollama is running: `ollama serve`

- **Node.js 18+ and npm** — For the UI (located in `./ui`)
  - Verify installation: `node --version` and `npm --version`

- **Git** — For version control
  - Verify installation: `git --version`

### Recommended

- **Make** — For running build targets. [Install Make](https://www.gnu.org/software/make/)
  - Verify installation: `make --version`

- **A code editor** — VS Code, GoLand, or your preferred editor

## Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/riteshmishra/llama-nest.git
cd llama-nest
```

### 2. Install Go Dependencies

```bash
go mod download
go mod tidy
```

### 3. Install UI Dependencies

```bash
cd ui
npm install
cd ..
```

### 4. Set Up Ollama

Start Ollama in a separate terminal:

```bash
ollama serve
```

By default, Ollama listens on `http://localhost:11434`.

### 5. Build the Project

Using Make:

```bash
make build
```

Or using Go directly:

```bash
go build -o llama-nest ./cmd/...
```

## Running llama-nest

### Start the Server

```bash
./llama-nest
```

Or using Make:

```bash
make run
```

The server will start and connect to your local Ollama instance.

### Running the UI

In a separate terminal:

```bash
cd ui
npm start
```

This will start the development server (typically on `http://localhost:3000`).

## Development Workflow

### Building

```bash
make build
```

### Running Tests

```bash
go test ./...
```

### Code Formatting

```bash
go fmt ./...
```

### Linting

```bash
go vet ./...
```

## Project Structure

- `cmd/` — Command-line interface entry points
- `internal/` — Core application logic and packages
- `ui/` — React-based web interface
- `docs/` — Documentation
- `Makefile` — Build automation

## Troubleshooting

### Ollama Connection Issues

If llama-nest can't connect to Ollama:

1. Ensure Ollama is running: `ollama serve`
2. Check the default port: Ollama runs on `http://localhost:11434`
3. Check logs for connection errors

### Go Version Mismatch

If you see Go version errors:

```bash
go version  # Should be 1.22 or higher
```

Update Go if necessary.

### UI Build Issues

If the UI fails to build:

```bash
cd ui
rm -rf node_modules package-lock.json
npm install
```

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Ollama Documentation](https://github.com/ollama/ollama)
- [Architecture Guide](./architecture.md)

## Getting Help

If you encounter issues:

1. Check the [Contributing Guide](../CONTRIBUTING.md)
2. Open an issue on GitHub with:
   - Your OS and version
   - Go version (`go version`)
   - Ollama version (`ollama --version`)
   - Steps to reproduce
