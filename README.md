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
| `JWT_SECRET` | **Required**; HS256 signing key for access tokens |
| `ENCRYPTER_KEY` | **Required** in production; 32-byte key as 64 hex characters |
| `DEPLOYMENT_ENVIRONMENT` | Set to `production` to enable fail-closed configuration checks |
| `IAM_ALLOW_INSECURE` | Local-only; skips token verification in competition routes |
| `IAM_MIGRATION_DRY_RUN` | When `true`, inspect pending migrations without applying changes |
| `IAM_MIGRATION_ROLLBACK` | Version of a migration to roll back on startup (e.g. `001`) |

### Database migrations

Migrations are applied on IAM startup. Each migration's data changes and its record are committed in one MongoDB transaction, and the record includes a checksum that detects edits to already-applied migrations. Production requires a MongoDB replica set (or mongos); a standalone `mongod` is only supported for local development and logs a warning. Migration `001` converts legacy policies in place, keeps a backup of the original value, and restores it on rollback.

### Competition

| Variable | Notes |
| --- | --- |
| `POSTGRES_DSN` | PostgreSQL connection string |
| `MONGO_URI` | MongoDB connection string for replays |
| `SIMULATION_URL` | gRPC address of the simulation engine (default `simulation:50051`) |
| `ALLOW_IN_MEMORY_FALLBACK` | Local-only escape hatch for missing databases |
| `IAM_URL` | Base URL used to verify bearer tokens (default `http://iam:7000`) |
| `MATCH_RESULT_SERVICE_EMAIL` | Optional; restrict result callbacks to this IAM service account |

### Gateway

`PORT`, `IAM_URL`, `COMPETITION_URL`, `REALTIME_URL`.

## Known gaps before production

- The gateway is exposed over plaintext HTTP without an Ingress/TLS.
- Project scoping for IAM/competition resources is still incomplete.
- Competition persistence and ELO scoring still need the fixes tracked in the repository issues.
- The `realtime` and `simulation` services referenced by compose are not defined here.
