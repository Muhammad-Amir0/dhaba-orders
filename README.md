# Dhaba Orders

Dhaba Orders is a local-development restaurant ordering system. Its five application services use React, Python, Go, Java, and PHP, with PostgreSQL for menu/order persistence and Redis for caching, jobs, and statistics.

## Run the complete system locally

Requirements: Docker Engine and Docker Compose v2.

```sh
cp .env.example .env
docker compose up
```

The sample credentials in `.env.example` are fake and for local testing only. Edit `.env` to change them. The compose file reads `.env.example` by default and lets `.env` override it. PostgreSQL initializes its schema and sample menu on the first start.

Open the web app at <http://localhost:8081>. The APIs are reachable locally at:

| Service | Local address |
| --- | --- |
| Menu API | <http://localhost:8000/api/menu> |
| Orders API | <http://localhost:8082/api/orders> |
| Stats API | <http://localhost:8083/api/stats> |
| PostgreSQL | `localhost:5432` |
| Redis | `localhost:6379` |

Compose uses the internal service DNS names `postgres` and `redis`; application containers do not depend on host IP addresses. Vite proxies the relative `/api/menu`, `/api/orders`, and `/api/stats` paths to their corresponding containers. The web client itself uses only relative API URLs.

To stop the system, press Ctrl+C. `docker compose down` stops and removes the containers but retains local database/Redis volumes. To reset local data, run `docker compose down -v` and start again.

## Shared API routes

- Menu: `GET /api/menu`, `/api/menu/healthz`, `/api/menu/livez`
- Orders: `POST /api/orders`, `GET /api/orders`, `GET /api/orders/{id}`, `/api/orders/info`, `/api/orders/healthz`, `/api/orders/livez`
- Stats: `GET /api/stats`, `/api/stats/info`, `/api/stats/healthz`, `/api/stats/livez`

The order request body is `{"items":[{"menu_item_id":1,"qty":2}]}`. Orders are queued in Redis; the kitchen worker progresses them from `new` to `cooking` and `ready`.

## Run and test each service locally

All services read database/Redis connection details from environment variables. `.env.example` documents the local values. Export its variables in your shell before running an API/worker outside Compose, or set equivalent values for your environment. Keep real credentials out of source control.

### Database

PostgreSQL 16 is initialized by `db/init.sql`, which creates the shared tables and seeds 12 Pakistani menu dishes. With PostgreSQL running, initialize an empty database with:

```sh
psql "postgresql://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB" \
  -f db/init.sql
```

### Menu API (Python 3.12)

```sh
cd menu-api
python3.12 -m venv .venv
. .venv/bin/activate
pip install -r requirements-dev.txt
pytest
ruff check .
uvicorn app:app --reload --port 8000
```

Ruff intentionally reports the unused `time` import planted for CI testing.

### Orders API (Go 1.22)

```sh
cd orders-api
go mod download
go test ./...
golangci-lint run
go run .
```

`golangci-lint` intentionally reports the unchecked JSON encoder error in the info handler planted for CI testing.

### Kitchen worker (Java 17, Maven)

```sh
cd kitchen-worker
mvn test
mvn checkstyle:check
mvn spring-boot:run
```

Set `KITCHEN_COOKING_TIME_SECONDS` to an integer from 5 through 10 to choose the simulated cooking duration.

### Stats API (PHP 8.3, Laravel 11)

```sh
cd stats-api
composer install
composer test
composer analyse
APP_ENV=production APP_KEY=local-development-only php artisan config:cache
php artisan serve --host 0.0.0.0 --port 8080
```

The intentionally incorrect PHPUnit expectation makes the full test suite fail by design; the other endpoint tests pass. The stats service reads directly from Redis and does not use a database for application data, sessions, or cache.

### Web app (Node.js)

```sh
cd web
npm ci
npm test
npm run lint
npm run build
npm run dev
```

For browser testing with the complete system and its Compose service-name API proxies, use `docker compose up` and open <http://localhost:8081>.

## Runtime configuration

The APIs and worker use environment variables for PostgreSQL and Redis hosts, ports, database/user/password, `APP_VERSION`, and `HOSTNAME`. The default version is `dev`. Compose sets dependency hosts to their service names and uses the fake local credentials in `.env.example`.
