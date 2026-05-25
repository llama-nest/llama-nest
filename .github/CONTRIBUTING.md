# Contributing to llama-nest

Thank you for contributing to `llama-nest`.

This project is exploring local-first conversational infrastructure for AI runtimes built on top of Ollama.

The long-term goal is not simply "memory for LLMs."

The goal is:

> portable conversational context infrastructure.

That means:
- inspectable local memory
- model interoperability
- transferable conversational state
- runtime observability
- local-first AI tooling

---

# Philosophy

Before contributing, it helps to understand the core design principles behind the project.

## 1. Local-first by default

`llama-nest` should work completely offline.

Avoid:
- cloud dependencies
- telemetry
- hosted APIs
- external services for core functionality

Users should fully own their conversational context.

---

## 2. Inspectable before autonomous

Raw captured context matters.

The project intentionally stores inspectable sessions/messages before attempting:
- memory extraction
- embeddings
- semantic graphs
- agent automation

We want users to understand and control what is being captured.

---

## 3. Model-agnostic design

`llama-nest` is not tied to:
- a specific model
- a specific vendor
- a specific runtime

The system should treat models as interchangeable execution engines while conversational context remains portable.

---

## 4. Simplicity over abstraction

Prefer:
- clear code
- explicit flows
- lightweight files
- understandable persistence

Avoid introducing heavy frameworks unless they provide significant value.

The current architecture intentionally favors:
- JSONL persistence
- small packages
- minimal dependencies
- inspectable storage

---

# Development setup

## Prerequisites

- Go 1.22+
- Ollama
- Node.js 20+ (for UI)

---

## Clone

```bash
git clone https://github.com/llama-nest/llama-nest.git
cd llama-nest
```

---

## Build

```bash
make build
```

---

## Start Ollama

```bash
ollama serve
```

---

## Initialize local storage

```bash
./bin/llama-nest init
```

---

## Start llama-nest

```bash
./bin/llama-nest start
```

---

## Run interactive chat

```bash
./bin/llama-nest run llama3.2
```

---

## Start UI

```bash
cd ui
npm install
npm run dev
```

Open:

```text
http://localhost:5173
```

---

# Repository structure

```text
cmd/
  llama-nest/        CLI entrypoint

internal/
  api/               HTTP API
  config/            runtime configuration
  db/                local persistence
  ollama/            Ollama helpers
  proxy/             request proxy layer
  transfer/          context transfer logic
  types/             shared types

ui/
  React/Vite frontend
```

---

# Common contribution areas

## Runtime monitoring

Planned:
- latency metrics
- throughput tracking
- tokens/sec
- model comparison graphs
- request instrumentation

---

## Transfer improvements

Planned:
- smarter context packing
- transfer compression
- semantic summarization
- selective memory transfer

---

## Storage evolution

Potential future work:
- sqlite-vec
- LanceDB
- semantic memory graphs
- encrypted local stores

---

## UI/UX

Potential improvements:
- monitoring dashboards
- transfer timelines
- memory inspection
- model comparison views
- conversational graph visualization

---

# Coding guidelines

## Go

Please:
- keep packages small
- avoid unnecessary abstractions
- prefer explicit code paths
- favor readability over cleverness

Run before submitting:

```bash
make build
```

---

## Frontend

Current UI stack:
- React
- Vite
- plain CSS

Please avoid:
- unnecessary UI frameworks
- excessive complexity
- over-engineered state management

The UI should remain lightweight and local-first.

---

# Pull requests

## Before opening a PR

Please:
- ensure the project builds
- keep PRs focused
- document architectural changes
- explain tradeoffs clearly

---

## PR titles

Recommended style:

```text
feat: add model latency metrics
fix: resolve stale PID handling
docs: improve architecture diagrams
```

---

## Good PRs usually include

- what changed
- why it changed
- architectural implications
- future considerations
- screenshots if UI-related

---

# Areas we are intentionally avoiding (for now)

The project is intentionally avoiding:
- cloud sync
- hosted APIs
- auth systems
- multi-user architecture
- enterprise abstractions
- premature orchestration complexity

The current focus is:
> making local conversational infrastructure understandable and portable.

---

# Roadmap direction

The project is currently evolving toward:

```text
local AI runtime
        +
portable conversational infrastructure
        +
runtime observability
```

Key future areas:
- monitoring
- distributed memory
- import/export portability
- semantic retrieval
- MCP integrations
- runtime orchestration
- model interoperability

---

# Questions / Discussions

If you're proposing:
- architectural changes
- storage migrations
- runtime abstractions
- major dependency additions

please open an issue/discussion first.

---

# License

By contributing, you agree that your contributions will be licensed under the repository license.
