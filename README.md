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

## Prerequisites

- **Ollama installed** ([Install Ollama](https://ollama.ai/))
- **Single Ollama instance** running on port `11434` (see below for important note)
- Go 1.22+ installed
- Node.js 18+ (for UI, optional)

## Setup

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

## Running llama-nest

`llama-nest` requires **three components running simultaneously**. Open three separate terminal windows and follow the steps in order:

⚠️ **All three terminals must be active at the same time for llama-nest to work.** If you skip any step or close a terminal, you'll get `connection refused` errors.

### Terminal 1: Ollama (must be running, but only one instance)

**IMPORTANT:** You must have **exactly one Ollama instance running** on port `11434`. Do not start multiple Ollama processes.

Check if Ollama is already running:

```bash
lsof -i :11434
```

If nothing is shown, start Ollama:

```bash
ollama serve
```

If the command fails with `address already in use`, Ollama is already running (either in another terminal or as a background service). In this case, you can skip this step and proceed to Terminal 2.

**On macOS:** Ollama may be running as a background service via the menu bar or LaunchAgent. You can:
- Keep it running and proceed to Terminal 2
- Or quit it first (if you want to restart it in a terminal)

### Terminal 2: Start llama-nest proxy

**CRITICAL:** You must start llama-nest server before running any commands in Terminal 3.

```bash
./bin/llama-nest start
```

This starts:
- **Proxy server** on `http://localhost:11435` (proxies requests to Ollama)
- **API server** on `http://localhost:8787` (manages local memory and context)

**Verify the server is running:**

```bash
lsof -i :11435
```

You should see `llama-nest` listening on port 11435. If nothing appears, the server failed to start—check the terminal output for errors.

### Terminal 3: Pull models and use llama-nest commands

Before running interactive chat, pull the model you want to use:

```bash
ollama pull llama3.2
```

Now you can run interactive chat, transfers, or other commands:

```bash
./bin/llama-nest run llama3.2
```

**Important:** The llama-nest server must be running (Terminal 2) before you use any commands.

---

# Using llama-nest

**Note:** All commands below assume the llama-nest server is running (`./bin/llama-nest start` in Terminal 2). If you see `connection refused` errors, start the server first.

## Interactive chat

```bash
./bin/llama-nest run llama3.2
```

Routes your chat through the llama-nest proxy so conversations and context are captured.

---

## Search captured context

```bash
./bin/llama-nest search "cookies"
```

Search all captured messages and sessions for relevant context.

---

## Generate catch-up brief

```bash
./bin/llama-nest catch-up
```

Generate a summary of recent conversation context to restore continuity.

---

## Transfer context between models

```bash
./bin/llama-nest transfer tinyllama
```

Move recent conversational context to another model, with automatic model pulling.

---

## View usage

```bash
./bin/llama-nest usage
```

View token usage across all captured sessions.

---

## Export local context

```bash
./bin/llama-nest export
```

Export all captured conversational state into a portable `.nest` bundle.

---

## Wipe local memory

```bash
./bin/llama-nest wipe --yes
```

Permanently delete all captured local memory and context.

---

## Stop llama-nest

```bash
./bin/llama-nest stop
```

Stop the proxy and API servers.

---

## Health check and diagnostics

```bash
./bin/llama-nest doctor
```

Comprehensive health check that verifies:
- Ollama is running and reachable
- Models are available locally
- llama-nest proxy is running
- llama-nest API is running

Shows exactly what to fix if something is missing.

---

# Troubleshooting

## "address already in use" on port 11434

```
Error: listen tcp 127.0.0.1:11434: bind: address already in use
```

**Solution:** Ollama is already running. Check:

```bash
lsof -i :11434
```

You have two options:

1. **Use the existing Ollama instance** (recommended) — Skip the `ollama serve` step and proceed directly to starting llama-nest
2. **Kill the existing process and restart:**
   ```bash
   kill -9 <PID>
   ollama serve
   ```

On macOS, Ollama may be running as a background service or in the menu bar. You can quit it from the menu or use `kill`.

---

## "connection refused" on port 11435

```
error: llama-nest server is not running

Fix: Run this in another terminal:
  ./bin/llama-nest start

Then come back and try again
```

**Solution:** The llama-nest server is not running.

Make sure you have started llama-nest in Terminal 2:

```bash
./bin/llama-nest start
```

All commands (like `./bin/llama-nest run llama3.2`) require the proxy server to be running.

Alternatively, run the health check to see what's missing:

```bash
./bin/llama-nest doctor
```

---

## Model not found / "pulling model"

If you see errors about a missing model, pull it first using Ollama:

```bash
ollama pull llama3.2
```

Make sure the model is available locally before running `./bin/llama-nest run <model>`.

---

## "llama-nest already running" but server won't start

```
error: llama-nest already running
```

This usually means the PID file is stale (the previous process crashed or was killed). The app now automatically detects this and cleans up stale PID files.

Try starting again:

```bash
./bin/llama-nest start
```

If it still fails, you can manually clean up:

```bash
rm ~/.llama-nest/llama-nest.pid
./bin/llama-nest start
```

---

## Verify your setup

Use the built-in health check:

```bash
./bin/llama-nest doctor
```

This automatically checks:
- ✓ Ollama is running and reachable
- ✓ Models are available locally
- ✓ llama-nest proxy is running
- ✓ llama-nest API is running

It will tell you exactly what to fix if anything is missing.

Alternatively, check manually:

```bash
# Check Ollama (port 11434)
lsof -i :11434

# Check llama-nest proxy (port 11435)
lsof -i :11435

# Check llama-nest API (port 8787)
lsof -i :8787
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
