# protobuf-ai-potato

Go gRPC AI chat backend with protobuf streaming. Docker-first so you can scale replicas later.

Optional simple RAG via Postgres + pgvector.

Live demo (grpcui): https://protobuf-ai-potato.onrender.com

## Layout

```
api/proto/chat/v1/chat.proto   # ChatService contract
cmd/server                     # gRPC server entrypoint
internal/llm                   # mock + openai-compatible (openai, groq) providers
internal/rag                   # chunk + embed + pgvector retrieve
internal/session               # in-memory session history
internal/server                # Chat RPC handlers
Dockerfile                     # buf generate + static binary (local chat)
Dockerfile.grpcui              # browser gRPC UI only
Dockerfile.render              # Render all-in-one: chat + grpcui
docker-compose.yml             # postgres + chat + grpcui (+ optional redis)
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
- Postgres + pgvector: `:5432`

Default provider is `mock` (no API key needed). RAG is on when `DATABASE_URL` is set (compose sets it by default).

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
# needs: go 1.22+, buf, postgres with pgvector
make run
```

## gRPC surface

- `Chat(ChatRequest) returns (stream ChatChunk)` — stream token deltas (injects RAG context when enabled)
- `IngestDocument` — chunk + embed + store a document for RAG
- `ListSessions` — list in-memory sessions
- standard gRPC health + reflection enabled

## Simple RAG

Flow:

1. Call `IngestDocument` with title + content
2. Text is split into chunks and embedded
3. Chunks are stored in Postgres (`vector(1536)`)
4. On `Chat`, the question is embedded, top-k similar chunks are retrieved, and added as a system message before the LLM call

Env:

```text
DATABASE_URL=postgres://potato:potato@localhost:5432/potato?sslmode=disable
EMBEDDING_PROVIDER=mock   # or openai
EMBEDDING_MODEL=text-embedding-3-small
RAG_TOP_K=4
```

Leave `DATABASE_URL` empty to disable RAG.

`EMBEDDING_PROVIDER=mock` works offline for demos. For real semantic search, set:

```text
EMBEDDING_PROVIDER=openai
OPENAI_API_KEY=sk-...
```

Try in grpcui:

1. `IngestDocument` with a short note, e.g. title `potato policy`, content `Employees get 12 potato days off per year.`
2. `Chat` with `How many potato days off do employees get?`

## Scaling path

1. Chat containers are mostly stateless: provider calls + in-memory history per process.
2. Document chunks live in Postgres (shared across replicas when RAG is enabled).
3. `docker compose --profile scale up --scale chat=2` starts Redis (history sharing not wired yet).
4. Next step: move `internal/session` to Postgres/Redis so any replica can continue a session.
5. Put a load balancer / ingress in front of `:50051` (or use Kubernetes Deployment + Service).

## Useful commands

```bash
make docker-up
make docker-down
make docker-scale
make generate
```

## Deploy on Render (UI in front)

Use `Dockerfile.render` so the public URL is grpcui (HTTP). gRPC stays inside the container.

1. Render → Web Service → Docker
2. Dockerfile path: `Dockerfile.render`
3. Env vars:

```text
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_...
GROQ_MODEL=llama-3.3-70b-versatile
```

RAG needs a managed Postgres with pgvector and `DATABASE_URL`. Without it, chat still works; ingest returns failed precondition.

Leave `PORT` alone (Render sets it). Do not point `GRPC_ADDR` at Render's public port — the entrypoint keeps gRPC on `127.0.0.1:50051`.

4. After deploy, open: `https://<your-service>.onrender.com`

Browser → grpcui (HTTP) → localhost gRPC chat inside the same container.
