# llama-nest

Local-first memory and context infrastructure for Ollama.

`llama-nest` is an experimental sidecar for local AI systems. It runs beside Ollama, captures local conversations through a proxy, stores an encrypted-by-design local event trail, and exposes a simple UI for inspecting sessions, memories, and generated catch-up briefs.

> Status: early experimental v0. The current implementation captures requests/responses, persists them locally in SQLite, exposes search, and generates basic catch-up briefs. Memory extraction, encryption-at-rest, graph relationships, and MCP integrations are planned next.

## Why this exists

Ollama made local model execution simple. But while testing models locally, especially on edge devices like NVIDIA Jetson Orin, context continuity becomes painful:

- conversations are fragmented across sessions
- model switching loses useful context
- local AI memory is hard to inspect
- retrieval and long-term continuity are not built in
- sensitive context needs local-first handling

`llama-nest` explores a small, inspectable memory layer for local AI.

## Quick start

Prerequisite: Ollama running locally on port `11434`.

```bash
ollama serve
```

In another terminal:

```bash
git clone https://github.com/riteshmishra/llama-nest.git
cd llama-nest
go mod tidy
go run ./cmd/llama-nest init
go run ./cmd/llama-nest start
```

The API/UI server runs on:

```text
http://localhost:8787
```

The Ollama proxy runs on:

```text
http://localhost:11435
```

Send traffic through the proxy:

```bash
curl http://localhost:11435/api/chat \
  -d '{"model":"llama3.2","messages":[{"role":"user","content":"Remember that llama-nest started from Jetson experiments."}],"stream":false}'
```

Then inspect:

```bash
go run ./cmd/llama-nest status
go run ./cmd/llama-nest search "Jetson"
go run ./cmd/llama-nest catch-up
```

## UI

The v0 UI is a Vite React app.

```bash
cd ui
npm install
npm run dev
```

Open:

```text
http://localhost:5173
```

The UI talks to the backend at `http://localhost:8787`.

## Commands

```bash
llama-nest init       # create local config and SQLite database
llama-nest start      # start API server and Ollama proxy
llama-nest status     # show local status
llama-nest search     # search captured context
llama-nest catch-up   # generate a local memory brief
llama-nest doctor     # validate local setup
```

## Design principles

- local-only by default
- no telemetry
- raw context first, derived memory second
- inspectable before autonomous
- model-agnostic memory
- user-controlled deletion/export planned

## Roadmap

- [ ] encrypted event store
- [ ] structured memory extraction schema
- [ ] entity graph builder
- [ ] sqlite-vec or LanceDB backend
- [ ] MCP server
- [ ] Tauri desktop app
- [ ] Homebrew tap
- [ ] Jetson install profile

## License

Apache-2.0 recommended for this project, but choose intentionally before publishing.
