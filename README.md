# CodeArena Platform services

Go services for the CodeArena control plane:

| Service | Default port | Description |
| --- | --- | --- |
| `gateway` | 8080 | HTTP entry point; proxies IAM, competition and realtime |
| `iam` | 7000 | Authentication, users, roles, projects and clients |
| `competition` | 7001 (HTTP), 50052 (gRPC) | Bots, matches, tournaments and scoreboards |

## Local development

This module uses a relative `replace` directive to `../codearena-core`, so clone `codearena-core` next to this repository.

```bash
cp .env.example .env
# edit .env with locally generated secrets
docker compose up --build
```

`JWT_SECRET`, `ENCRYPTER_KEY` and `POSTGRES_PASSWORD` are required; compose refuses to start without them. The `competition` service fails fast when its databases are unavailable. Set `ALLOW_IN_MEMORY_FALLBACK=true` only for local development without databases.

The gateway exposes IAM under `/api/v1/iam/`. The legacy `/api/v1/auth/` prefix is rewritten to `/api/v1/iam/auth/`, and `/api/v1/users/` to `/api/v1/iam/user/v1/`.

## Configuration

### IAM

| Variable | Notes |
| --- | --- |
| `SERVER_HOST`, `SERVER_PORT` | Bind address; defaults to `:3000` |
| `DB_HOST` | MongoDB connection URI |
| `REDIS_HOST` | Redis `host:port` |
| `JWT_SECRET` | **Required**; signing key for access tokens |
| `ENCRYPTER_KEY` | **Required**; 32-byte key as 64 hex characters |
| `IAM_ALLOW_INSECURE` | Only for local development; when `true`, authenticated routes pass through if no IAM client is configured |

### Competition

| Variable | Notes |
| --- | --- |
| `POSTGRES_DSN` | PostgreSQL connection string |
| `MONGO_URI` | MongoDB connection string for replays |
| `SIMULATION_URL` | gRPC address of the simulation engine (default `simulation:50051`) |
| `ALLOW_IN_MEMORY_FALLBACK` | Local-only escape hatch for missing databases |

### Gateway

`PORT`, `IAM_URL`, `COMPETITION_URL`, `REALTIME_URL`.

## Known gaps before production

- The gateway is exposed over plaintext HTTP without an Ingress/TLS.
- No authentication on competition mutation endpoints; `X-User-ID` is currently trusted and must be replaced with verified token claims.
- Several IAM authorization issues remain (project scoping, OAuth code/client binding, PKCE validation, token cache invalidation).
- The `realtime` and `simulation` services referenced by compose are not defined here.
- OAuth token refresh, Mongo field names and Postgres schema need the fixes tracked in the repository issues.
