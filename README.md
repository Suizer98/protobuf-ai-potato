# protobuf-ai-potato

Go gRPC AI chat backend with protobuf streaming. Docker-first so you can scale replicas later.

## Layout

```
api/proto/chat/v1/chat.proto   # ChatService contract
cmd/server                     # gRPC server entrypoint
internal/llm                   # mock + openai-compatible (openai, groq) providers
internal/session               # in-memory session history
internal/server                # Chat RPC handlers
Dockerfile                     # buf generate + static binary
docker-compose.yml             # chat service (+ optional redis)
```

## Quick start

### Docker

```bash
cp .env.example .env
docker compose up --build
```

Server listens on `:50051`. Default provider is `mock` (no API key needed).

### Groq

```bash
# .env
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_...
GROQ_MODEL=llama-3.3-70b-versatile
docker compose up --build
```

### OpenAI

```bash
# .env
LLM_PROVIDER=openai
OPENAI_API_KEY=sk-...
docker compose up --build
```

### Local without Docker

```bash
# needs: go 1.22+, buf
make run
```

## gRPC surface

- `Chat(ChatRequest) returns (stream ChatChunk)` — stream token deltas
- `ListSessions` — list in-memory sessions
- standard gRPC health + reflection enabled

## Scaling path

1. Chat containers are mostly stateless: provider calls + in-memory history per process.
2. `docker compose --profile scale up --scale chat=2` starts Redis (history sharing not wired yet).
3. Next step: move `internal/session` to Redis so any replica can continue a session.
4. Put a load balancer / ingress in front of `:50051` (or use Kubernetes Deployment + Service).

## Useful commands

```bash
make docker-up
make docker-down
make docker-scale
make generate
```
