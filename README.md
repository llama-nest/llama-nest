# llama-nest

Local-first memory and context infrastructure for Ollama.

`llama-nest` is an experimental runtime wrapper around Ollama that adds:

- inspectable local memory
- conversational context persistence
- model-to-model context transfer
- token usage tracking
- local search
- catch-up briefs
- runtime monitoring foundations

The goal is simple:

> context should outlive models.

`llama-nest` sits between your applications and Ollama, capturing and structuring conversational context locally while remaining fully model-agnostic.

---

# What problem does this solve?

Ollama makes running local models easy.

But local AI workflows still have major gaps:

- conversations become fragmented
- switching models loses useful context
- there is no inspectable memory layer
- usage and performance are difficult to observe
- context continuity is tied to a single model session
- local AI tooling lacks portable conversational infrastructure

`llama-nest` explores a local-first memory and portability layer for local AI runtimes.

---

# How llama-nest works

Instead of applications talking directly to Ollama:

```text
app ─────► ollama
```

Applications talk to `llama-nest`, which proxies requests into Ollama:

```text
                ┌────────────────────────────┐
                │         Your App           │
                │ CLI / scripts / agents/UI │
                └─────────────┬──────────────┘
                              │
                              │ requests
                              ▼
                ┌────────────────────────────┐
                │         llama-nest         │
                │                            │
                │  local proxy + memory      │
                │                            │
                │  localhost:11435           │
                └─────────────┬──────────────┘
                              │
                              │ proxied requests
                              ▼
                ┌────────────────────────────┐
                │           Ollama           │
                │      localhost:11434      │
                └────────────────────────────┘
```

Because `llama-nest` sits in the middle, it can:

- capture conversations
- persist local memory
- track token usage
- monitor model performance
- search prior context
- transfer context between models
- export/import conversational state

---

# Core architecture

```text
                ┌─────────────────┐
                │     Ollama      │
                │    localhost    │
                │      11434      │
                └────────┬────────┘
                         │
                         │
                ┌────────▼────────┐
                │   llama-nest    │
                │      proxy      │
                │      11435      │
                └────────┬────────┘
                         │
         ┌───────────────┼─────────────────────┐
         │               │                     │
         ▼               ▼                     ▼
    sessions         messages              transfers
         │
         ├─────────────────────────────────────┐
         │                                     │
         ▼                                     ▼
   local search                          usage tracking
         │                                     │
         ▼                                     ▼
   catch-up briefs                    future monitoring
```

---

# Features

## Local memory capture

All conversations passing through the proxy are captured locally.

`llama-nest` currently stores:

- sessions
- messages
- transfers
- token usage
- generated catch-up context

Everything remains local-first.

No telemetry.

No cloud dependency.

---

## Interactive local chat

Instead of using:

```bash
ollama run llama3.2
```

you can use:

```bash
llama-nest run llama3.2
```

This routes the chat through the `llama-nest` proxy so context, usage, and future transfers are captured automatically.

---

## Catch-up briefs

Generate a brief of recent conversational context:

```bash
llama-nest catch-up
```

Useful for:

- re-entering long sessions
- restoring conversational continuity
- summarizing recent local memory

---

## Context transfer between models

Move recent conversational context between local models.

Example:

```bash
llama-nest transfer qwen2.5:0.5b
```

`transfer`:

- checks if the target model exists locally
- pulls it automatically if missing
- builds a local context pack
- transfers recent conversational context
- asks the target model to acknowledge the session

This allows conversations to continue across models without losing continuity.

Example:

```text
Checking model tinyllama...
Building context pack...
Transferring session...
tinyllama is caught up.
```

Transfers are persisted locally and visible in the UI.

---

## Token usage tracking

`llama-nest` tracks:

- prompt tokens
- completion tokens
- total tokens
- usage by model

Example:

```bash
llama-nest usage
```

Future monitoring support will include:

- latency
- tokens/sec
- request throughput
- model performance comparisons
- runtime metrics

---

## Export local context

Export captured conversational state into a portable `.nest` bundle:

```bash
llama-nest export
```

Planned future support:

```bash
llama-nest import
```

The long-term goal is portable conversational infrastructure across machines and models.

---

# Quick start

Prerequisite:

- Ollama installed
- Ollama running locally on port `11434`

Start Ollama:

```bash
ollama serve
```

Clone and build:

```bash
git clone https://github.com/riteshmishra/llama-nest.git
cd llama-nest

make build
```

Initialize local storage:

```bash
./bin/llama-nest init
```

Start the local sidecar:

```bash
./bin/llama-nest start
```

The API server runs on:

```text
http://localhost:8787
```

The proxy runs on:

```text
http://localhost:11435
```

---

# Using llama-nest

## Interactive chat

```bash
./bin/llama-nest run llama3.2
```

---

## Search captured context

```bash
./bin/llama-nest search "cookies"
```

---

## Generate catch-up brief

```bash
./bin/llama-nest catch-up
```

---

## Transfer context between models

```bash
./bin/llama-nest transfer tinyllama
```

---

## View usage

```bash
./bin/llama-nest usage
```

---

## Export local context

```bash
./bin/llama-nest export
```

---

## Wipe local memory

```bash
./bin/llama-nest wipe --yes
```

---

## Stop llama-nest

```bash
./bin/llama-nest stop
```

---

# UI

The UI is a lightweight Vite React application.

Start it separately:

```bash
cd ui
npm install
npm run dev
```

Open:

```text
http://localhost:5173
```

The UI currently supports:

- session inspection
- message browsing
- local search
- transfer history
- token usage tracking
- catch-up briefs

Future UI support:

- model monitoring
- performance charts
- latency analysis
- throughput graphs
- model comparisons

---

# Commands

```bash
llama-nest init             # initialize local config and storage
llama-nest start            # start local proxy + API
llama-nest stop             # stop running llama-nest services

llama-nest run MODEL        # interactive chat through llama-nest
llama-nest transfer MODEL   # transfer recent context to another model

llama-nest status           # show local status
llama-nest usage            # show token usage
llama-nest search QUERY     # search local context
llama-nest catch-up         # generate memory brief

llama-nest export           # export local context bundle
llama-nest wipe --yes       # delete captured local memory

llama-nest doctor           # validate local setup
```

---

# Data storage

Current local storage uses JSONL-backed local persistence inside:

```text
~/.llama-nest/
```

Current files include:

```text
sessions.jsonl
messages.jsonl
transfers.jsonl
usage.jsonl
```

Future versions may support:

- sqlite-vec
- LanceDB
- semantic memory graphs
- encrypted local stores

---

# Design principles

- local-first by default
- no telemetry
- inspectable before autonomous
- model-agnostic conversational context
- raw context before derived memory
- portability over lock-in
- user-controlled memory lifecycle

---

# Roadmap

- [ ] latency + throughput monitoring
- [ ] Grafana-style monitoring dashboard
- [ ] encrypted local event store
- [ ] structured memory extraction
- [ ] semantic memory graph
- [ ] sqlite-vec / vector backend
- [ ] model routing engine
- [ ] import `.nest` bundles
- [ ] shared memory across devices
- [ ] MCP server
- [ ] Tauri desktop app
- [ ] Homebrew tap
- [ ] Docker image
---

# Development

```bash
make fmt
make build
```

---

# License

Apache-2.0
