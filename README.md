# protobuf-ai-potato

Go gRPC AI chat backend with protobuf streaming. Docker-first so you can scale replicas later.

Live demo (grpcui): https://protobuf-ai-potato.onrender.com

## Layout

```
api/proto/chat/v1/chat.proto   # ChatService contract
cmd/server                     # gRPC server entrypoint
internal/llm                   # mock + openai-compatible (openai, groq) providers
internal/session               # in-memory session history
internal/server                # Chat RPC handlers
Dockerfile                     # buf generate + static binary (local chat)
Dockerfile.grpcui              # browser gRPC UI only
Dockerfile.render              # Render all-in-one: chat + grpcui
docker-compose.yml             # chat + grpcui (+ optional redis)
scripts/render-entrypoint.sh   # starts gRPC then grpcui

```

## Quick start

### Docker

```bash
cp .env.example .env
docker compose up --build
```

- gRPC server: `:50051`
- Browser UI (grpcui): [http://localhost:8080](http://localhost:8080)

Default provider is `mock` (no API key needed).

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

## Deploy on Render (UI in front)

Use `Dockerfile.render` so the public URL is **grpcui** (HTTP). gRPC stays inside the container.

1. Render → Web Service → Docker
2. Dockerfile path: `Dockerfile.render`
3. Env vars:

```text
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_...
GROQ_MODEL=llama-3.3-70b-versatile
```

Leave `PORT` alone (Render sets it). Do **not** point `GRPC_ADDR` at Render's public port — the entrypoint keeps gRPC on `127.0.0.1:50051`.

4. After deploy, open: `https://<your-service>.onrender.com`

Browser → grpcui (HTTP) → localhost gRPC chat inside the same container.
