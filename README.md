# CampaignFlow

CampaignFlow is a portfolio SaaS-style application for campaign management, bid simulation, and metrics reporting.

The project is currently in early backend foundation work. The planned stack is a Go API, PostgreSQL, Redis, and a React/TypeScript dashboard. Product and architecture details live in [docs/implementation-plan.md](docs/implementation-plan.md).

## Current Status

- Go backend foundation with configuration, structured logging, graceful shutdown, and `GET /healthz`
- Docker Compose services for PostgreSQL and Redis
- PostgreSQL migration scripts using `golang-migrate/migrate`
- Planned persistence layer with `pgx`

## Repository Layout

```text
campaignflow/
  backend/
    cmd/api/                 Go API entrypoint
    internal/                Backend packages
      db/                    PostgreSQL connection helpers
      migrations/            SQL migration files
  docs/
    implementation-plan.md   Product and architecture plan
  scripts/                   Local/deploy helper scripts
  docker-compose.yml         Local PostgreSQL and Redis services
```

## Prerequisites

- Go
- Docker and Docker Compose
- Bash for scripts in `scripts/`
- `golang-migrate/migrate` CLI

Install the migration CLI:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Make sure your Go binary directory is on `PATH`.

## Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

The default local database URL is:

```env
DATABASE_URL=postgres://campaignflow:campaignflow@localhost:5432/campaignflow?sslmode=disable
```

## Local Services

Start PostgreSQL and Redis:

```bash
docker compose up -d postgres redis
```

Stop services:

```bash
docker compose down
```

Remove service volumes when you want a clean database/cache:

```bash
docker compose down -v
```

## Database Migrations

Migration files live in `backend/internal/migrations`.

Make the scripts executable on Linux:

```bash
chmod +x scripts/*.sh
```

Create a migration:

```bash
./scripts/migrate-create.sh create_initial_schema
```

Apply pending migrations:

```bash
./scripts/migrate-up.sh
```

Show current migration version:

```bash
./scripts/migrate-version.sh
```

Roll back one migration:

```bash
./scripts/migrate-down.sh 1
```

Force a migration version after manually resolving a dirty migration state:

```bash
./scripts/migrate-force.sh <version>
```

The scripts use `DATABASE_URL` from the environment. For local development, they also fall back to `.env` if `DATABASE_URL` is not already set.

## Run The API

From the repository root:

```bash
go run ./backend/cmd/api
```

Health check:

```bash
curl -i http://127.0.0.1:8080/healthz
```

Expected response: `204 No Content`.

## Development Notes

- Keep implementation decisions aligned with [docs/implementation-plan.md](docs/implementation-plan.md).
- Keep this README updated when setup steps, scripts, services, or major capabilities change.
- Keep migrations explicit and reviewable with paired `up.sql` and `down.sql` files.
