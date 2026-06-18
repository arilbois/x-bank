# x-bank

Modular-monolith Go backend for the x-bank platform. Aggregates
articles from three content categories (`sambatWarga`, `persibWay`,
`bytmod`), scores them, persists to PostgreSQL, and exposes a JSON HTTP
API protected by JWT auth.

## Stack

- Go 1.25.5
- Gin (`github.com/gin-gonic/gin`)
- GORM + PostgreSQL driver
- Colly (HTML scraping)
- robfig/cron/v3 (scheduled jobs)
- golang-jwt/jwt v5
- bcrypt password hashing
- godotenv (`.env` loading)

No Docker / Redis / Kafka / RabbitMQ / Elasticsearch.

## Quick start

```bash
# 1. Copy and edit env
cp .env.example .env

# 2. Fetch deps and build
make tidy
make build

# 3. Run (DB must be reachable)
./bin/server
```

Server listens on `APP_PORT` (default `8080`). On first run, an admin
user is seeded from `ADMIN_USERNAME` / `ADMIN_PASSWORD`.

## Layout

```
cmd/server/         main entrypoint
internal/config     env loading
internal/database   GORM + AutoMigrate
internal/models     User, Article, ArticleAnalysis, ScrapeLog
internal/repositories   CRUD + model-specific helpers
internal/services/auth      bcrypt + JWT
internal/services/scorer    rule-based scoring (0-100)
internal/services/scraper   9 Colly scrapers + orchestrator
internal/services/ai        OpenAI-compatible provider + analyzer
internal/handlers    Gin HTTP handlers
internal/middleware  auth, cors, logger
internal/scheduler   robfig/cron jobs
internal/routes      route wiring
tests/               unit tests
```

## API

All endpoints (except `/health` and `/auth/login`) require a Bearer
token. Admin-only routes require role `admin`.

| Method | Path                          | Auth    | Description                       |
|--------|-------------------------------|---------|-----------------------------------|
| GET    | `/health`                     | public  | healthcheck                       |
| POST   | `/auth/login`                 | public  | exchange username/password for JWT |
| GET    | `/articles`                   | auth    | paginated list, filters via query |
| GET    | `/articles/:id`               | auth    | fetch one article                 |
| GET    | `/trending`                   | auth    | top-scored articles               |
| GET    | `/analysis/:id`               | auth    | fetch one analysis                |
| GET    | `/articles/:id/analysis`      | auth    | latest analysis for an article    |
| POST   | `/scrape/run`                 | admin   | trigger a scrape run              |

### Article list query params

`category` (one of `sambatWarga|persibWay|bytmod`), `source`, `status`,
`sort` (`recent|score|oldest`), `page` (default 1), `limit` (default 20,
max 100).

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin12345"}'
```

Returns `{ "token": "...", "user": {...} }`. Use the token as
`Authorization: Bearer <token>`.

## Scraper

Nine Colly-based scrapers are registered out of the box:

- `sambatWarga`: cnbc, detik, kompas
- `persibWay`:   persib_official, simamaung, bolanet
- `bytmod`:      hackernews, github_trending, devto

Each scraper is best-effort: failures are logged and surfaced through
`scrape_logs` but do not bring down the whole run.

## Scoring (rule-based)

| Factor          | Weight                                 |
|-----------------|----------------------------------------|
| keyword match   | +30 if title/excerpt hits niche kw     |
| source weight   | 8–20 per source (e.g. persib=20, hn=15)|
| recency         | +20 (<1h), +15 (<6h), +10 (<24h), +5 (<72h) |

Final score is clamped to 0–100.

## AI analyzer

The AI provider is **configurable**: set `AI_BASE_URL`, `AI_API_KEY`,
`AI_MODEL` to point at any OpenAI-compatible `/chat/completions`
endpoint. The provider never receives the article body — only the title
and excerpt (truncated to 600 chars).

## Scheduled jobs

| Spec             | Category      |
|------------------|---------------|
| `0 */15 * * * *` | sambatWarga   |
| `0 */10 * * * *` | persibWay     |
| `0 */30 * * * *` | bytmod        |

(All three fire on second `0` of the minute, every N minutes.)

## Tests

```bash
make test
```

Covers scoring logic, password hash/verify, JWT round-trip, and the AI
prompt invariant (no full body ever sent to the provider).
