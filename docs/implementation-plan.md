# CampaignFlow Implementation Plan

## Summary

CampaignFlow will be a greenfield monorepo full-stack SaaS-style portfolio app.

Chosen defaults:

- Repository: monorepo with `backend/`, `frontend/`, `deploy/`, and root Docker Compose
- Backend: Go `net/http`, `pgx`, handwritten SQL, PostgreSQL migrations
- Auth: JWT access token in httpOnly cookie, DB-backed refresh session, CSRF header for unsafe requests
- Frontend: React + TypeScript + Vite, React Router, TanStack Query, React Hook Form, Zod, shadcn/ui, Recharts
- Redis: active campaign cache, login/bid rate limits, pacing/spend counters
- MVP focus: auth, organizations, campaign CRUD, bid simulation, metrics dashboard

## Project Progress

This document is a living plan. Update it as milestones are completed, scope changes, or better implementation choices emerge during the project.

Current status: planning.

## Recommended Repository Structure

```text
campaignflow/
  docker-compose.yml
  .env.example
  README.md
  docs/
    api.md
    architecture.md
    milestones.md

  backend/
    go.mod
    cmd/api/main.go
    migrations/
    internal/
      config/
      db/
      redis/
      logger/
      server/
      middleware/
      auth/
      organizations/
      users/
      campaigns/
      bidding/
      metrics/
      httperror/
      testutil/

  frontend/
    package.json
    index.html
    vite.config.ts
    tailwind.config.ts
    components.json
    src/
      main.tsx
      app/
        router.tsx
        providers.tsx
      layouts/
        dashboard-layout.tsx
        auth-layout.tsx
      pages/
        login.tsx
        register.tsx
        dashboard.tsx
        campaigns.tsx
        campaign-detail.tsx
        campaign-form.tsx
        bid-simulator.tsx
        analytics.tsx
        ml-insights.tsx
      features/
        auth/
        campaigns/
        bidding/
        metrics/
      components/
        ui/
        navigation/
        charts/
      lib/
        api-client.ts
        query-client.ts
        csrf.ts
        utils.ts
```

## Backend Implementation Phases

| Phase | Goal | Files To Create | Concepts Learned | Success Criteria |
|---|---|---|---|---|
| 1. Backend foundation | Start a minimal Go API with config, logging, health checks, graceful shutdown | `backend/cmd/api/main.go`, `internal/config`, `internal/logger`, `internal/server` | env config, `context.Context`, `slog`, graceful shutdown | `GET /healthz` works locally and shuts down cleanly |
| 2. Docker + database | Run Postgres, Redis, and backend dependencies locally | `docker-compose.yml`, `.env.example`, `backend/migrations` | service containers, env wiring, migrations | `docker compose up` starts Postgres/Redis and migrations apply |
| 3. Persistence layer | Add `pgxpool`, migration runner, store patterns | `internal/db`, feature store files | connection pools, transactions, SQL boundaries | API can connect to DB and fail readiness if DB is unavailable |
| 4. Auth | Register, login, logout, refresh, `me`, auth middleware | `internal/auth`, `internal/users`, middleware | password hashing, JWT cookies, refresh sessions, CSRF, protected routes | User can register, login, refresh, logout, and access protected route |
| 5. Organizations | Create organization during registration and scope all data to org | `internal/organizations` | tenant scoping, ownership boundaries | New user belongs to one org; protected APIs resolve `user_id` and `org_id` |
| 6. Campaign CRUD | Create/list/detail/update/pause/resume/delete campaigns | `internal/campaigns` | REST design, validation, service/store separation | Full campaign lifecycle works with org isolation |
| 7. Redis foundation | Add Redis client, health check, rate limiter helper, cache helper | `internal/redis`, `internal/middleware/ratelimit` | TTLs, fixed-window limits, cache invalidation | Login and bid endpoints can be rate limited |
| 8. Bid simulation | Accept fake bid requests and store decisions | `internal/bidding` | targeting, budget checks, latency measurement, decision reasons | Endpoint returns bid/no-bid and persists decision |
| 9. Metrics | Report spend, impressions, clicks, conversions, CTR, CVR, avg bid, latency | `internal/metrics` | SQL aggregation, derived metrics, date ranges | Dashboard APIs return accurate metrics from stored events |
| 10. Hardening | Add request IDs, consistent errors, validation, tests, seed data | shared middleware + test helpers | observability, API consistency, integration testing | MVP is demoable end-to-end with seeded sample data |

## Frontend Implementation Phases

| Phase | Goal | Files To Create | Concepts Learned | Success Criteria |
|---|---|---|---|---|
| 1. Frontend foundation | Scaffold Vite React TS, Tailwind, shadcn/ui | `frontend/src/main.tsx`, `app/providers.tsx`, `components/ui` | Vite, Tailwind, component system | App loads with base theme and providers |
| 2. Routing + layout | Add auth layout, dashboard layout, sidebar, top nav | `app/router.tsx`, `layouts/*`, `components/navigation` | React Router, protected layouts | Auth pages and dashboard shell route correctly |
| 3. API client | Add fetch wrapper with cookies, CSRF, TanStack Query | `lib/api-client.ts`, `lib/query-client.ts`, `lib/csrf.ts` | query caching, mutations, error states | Frontend can call `/auth/me` and handle auth failures |
| 4. Auth UI | Login/register forms with Zod validation | `pages/login.tsx`, `pages/register.tsx`, `features/auth` | RHF, Zod, form errors | User can register/login/logout from UI |
| 5. Campaign UI | Campaign table, create/edit form, detail page, pause/resume/delete | `features/campaigns`, campaign pages | data tables, optimistic/refetch flows, form schemas | Full campaign lifecycle works from dashboard |
| 6. Bid simulator | Build fake bid request form and decision result panel | `pages/bid-simulator.tsx`, `features/bidding` | simulator UX, request/response inspection | User can submit bid requests and see bid/no-bid reason |
| 7. Analytics | Overview cards and Recharts charts | `pages/dashboard.tsx`, `pages/analytics.tsx`, `components/charts` | derived metrics, chart states, empty/loading/error UI | Metrics render for org and campaign detail |
| 8. Polish | Responsive layout, loading skeletons, toasts, empty states | shared UI components | SaaS dashboard ergonomics | App feels coherent and demo-ready |

Frontend visual direction:

- Clean operational SaaS dashboard, not a marketing landing page
- Sidebar for primary navigation: Dashboard, Campaigns, Bid Simulator, Analytics, ML Insights
- Use metric cards only where they represent real KPIs
- Prefer dense tables, clear filters, and restrained color
- Use one accent color for primary actions/status emphasis

## Database Schema Plan

Use UUID primary keys, UTC timestamps, integer cents for money, and soft delete for campaigns.

Core tables:

- `organizations`: `id`, `name`, `created_at`, `updated_at`
- `users`: `id`, `organization_id`, `email`, `email_normalized`, `password_hash`, `role`, `created_at`, `updated_at`
- `auth_sessions`: `id`, `user_id`, `refresh_token_hash`, `user_agent`, `ip_address`, `expires_at`, `revoked_at`, `created_at`
- `campaigns`: `id`, `organization_id`, `name`, `status`, `daily_budget_cents`, `total_budget_cents`, `max_bid_cents`, `targeting_rules jsonb`, `start_at`, `end_at`, `created_at`, `updated_at`, `deleted_at`
- `bid_requests`: `id`, `organization_id`, `auction_id`, `country`, `device`, `placement`, `segments jsonb`, `floor_price_cents`, `created_at`
- `bid_decisions`: `id`, `organization_id`, `bid_request_id`, `campaign_id nullable`, `decision`, `reason`, `bid_price_cents nullable`, `latency_ms`, `created_at`
- `campaign_events`: `id`, `organization_id`, `campaign_id`, `bid_decision_id nullable`, `event_type`, `spend_cents`, `created_at`

Important constraints:

- Campaign status: `draft`, `active`, `paused`, `archived`
- Event type: `impression`, `click`, `conversion`
- Unique normalized user email
- Index campaign queries by `organization_id`, `status`, and date window
- Index metrics by `campaign_id, created_at` and `organization_id, created_at`

## API Route Plan

All routes live under `/api/v1`.

Auth:

- `POST /auth/register`: create organization and first user
- `POST /auth/login`: set access, refresh, and CSRF cookies
- `POST /auth/logout`: revoke refresh session and clear cookies
- `POST /auth/refresh`: rotate refresh session and issue new access cookie
- `GET /auth/me`: return current user and organization

Campaigns:

- `GET /campaigns?status=&page=&page_size=`
- `POST /campaigns`
- `GET /campaigns/{id}`
- `PATCH /campaigns/{id}`
- `POST /campaigns/{id}/pause`
- `POST /campaigns/{id}/resume`
- `DELETE /campaigns/{id}`

Bidding:

- `POST /bid-requests`: submit fake auction request
- `POST /bid-decisions/{id}/events`: record impression, click, or conversion outcome

Metrics:

- `GET /metrics/overview?from=&to=`
- `GET /campaigns/{id}/metrics?from=&to=`

System:

- `GET /healthz`
- `GET /readyz`

Standard response rules:

- JSON only
- Cookie auth with `credentials: include`
- Unsafe requests require `X-CSRF-Token`
- Errors use `{ "error": { "code": "...", "message": "..." } }`

## Redis Integration Plan

Use Redis in three focused places.

Active campaign cache:

- Key: `active_campaigns:{organization_id}`
- TTL: 60 seconds
- Invalidate on campaign create/update/pause/resume/delete
- Fallback to PostgreSQL if cache misses or Redis is unavailable

Rate limits:

- Login: `rl:login:{ip}:{email}`, 5 attempts per 15 minutes
- Bid endpoint: `rl:bid:{organization_id}:{ip}`, 120 requests per minute
- Use fixed-window counters for MVP

Pacing counters:

- Daily spend: `pace:{campaign_id}:{yyyymmdd}:spend_cents`
- Total spend: `pace:{campaign_id}:total_spend_cents`
- Increment on impression event
- Rebuild from PostgreSQL aggregates if Redis key is missing

## Suggested Milestone Order

1. Backend health check, config, logging, graceful shutdown
2. Docker Compose with Postgres and Redis
3. Migrations and database connection
4. Auth and organization creation
5. Campaign CRUD backend
6. Frontend scaffold, routing, dashboard shell
7. Frontend auth flow
8. Campaign table/forms/detail UI
9. Redis rate limiting and campaign cache
10. Bid simulation backend and frontend page
11. Metrics backend and dashboard charts
12. Polish, tests, README, seed data, demo script

## What To Build First

Build the backend foundation first:

1. `docker-compose.yml` with Postgres and Redis
2. Go API with `/healthz` and `/readyz`
3. Config loading from env
4. `pgxpool` connection
5. First migration for `organizations` and `users`

Do not start with the frontend. Auth, campaign ownership, and API shape will drive the frontend architecture.

## Testing Strategy

Backend:

- Unit test auth token handling, password hashing, campaign validation, targeting matching, bid decision logic
- Handler test protected routes with `httptest`
- Store integration tests against a test Postgres database
- Redis tests for rate-limit behavior using a test Redis instance or replaceable interface
- Migration smoke test from empty DB to latest schema

Frontend:

- Zod schema tests for auth and campaign forms
- React Testing Library tests for login form, campaign form, protected routing
- MSW mocks for API behavior
- Playwright smoke test after MVP: register, create campaign, simulate bid, view metrics

Manual demo acceptance:

- A new user can register and gets an organization
- User can create, update, pause, resume, and delete campaigns
- Bid simulator returns clear bid/no-bid reasons
- Metrics update after simulated impressions/clicks/conversions
- Login and bid rate limits can be demonstrated

## Stretch Goals After MVP

- Python FastAPI ML scoring service behind `/score`
- Async event ingestion worker
- Materialized metric rollups
- Campaign pacing algorithm beyond simple spend caps
- OpenAPI spec and generated frontend types
- Admin role and multi-user organization invites
- WebSocket/SSE live bid stream
- Audit logs for campaign changes
- Background jobs for campaign status transitions
- Deployment to Fly.io, Render, Railway, or AWS
