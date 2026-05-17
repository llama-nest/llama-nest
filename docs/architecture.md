# Architecture

```text
Client / App / CLI
      |
      v
llama-nest proxy :11435
      |
      v
Ollama :11434

Capture pipeline:
request/response -> event log -> search index -> catch-up brief -> UI/API
```

## Current v0

- Captures Ollama proxy traffic
- Stores sessions and messages in SQLite
- Provides API endpoints for UI and CLI
- Provides simple search and catch-up brief generation

## Intended production model

- Append-only encrypted raw event store
- Derived memory index
- Vector search backend
- Entity graph backend
- Local and optional cloud extractors
- MCP-compatible context server
