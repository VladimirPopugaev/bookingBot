# Add Postgres persistence for analyzed site results

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into the repository root and this document is maintained in accordance with it.

## Purpose / Big Picture

After this change, the service will have a dedicated PostgreSQL infrastructure repository that is responsible for connection pool creation, configuration validation, health checking, transaction entry points, and clean shutdown. This gives the project a reusable database foundation that future SQL repositories can consume without introducing an ORM. You can see the change working by starting PostgreSQL, running the service, and observing that startup succeeds only when the configured database is reachable.

## Progress

- [x] 2026-05-15 13:45Z Reviewed repository rules, current wiring in `cmd/run.go`, `internal/app/app.go`, `internal/usecase`, and existing config/docker files.
- [x] 2026-05-15 13:45Z Chosen implementation shape: shared Postgres manager in `internal/repository/postgres`, concrete repository in `internal/repository/postgres/siteanalysis`, persistence triggered from `Usecase.AnalyzeSite`.
- [x] 2026-05-15 13:45Z Added an ExecPlan file for the work and captured the intended schema, dependencies, and validation flow.
- [x] 2026-05-15 14:05Z Implemented `internal/repository/postgres` manager with config validation, `pgxpool` ownership, `Ping`, `Close`, `Pool`, `Builder`, and `BeginTx`.
- [x] 2026-05-15 14:05Z Wired the shared Postgres manager into `internal/app/app.go` startup and shutdown without changing use case or HTTP behavior.
- [x] 2026-05-15 14:05Z Added focused unit tests for config validation, placeholder formatting, and nil-pool transaction behavior.
- [x] 2026-05-15 14:06Z Ran `gofmt`, `go build ./...`, and `go test ./...` successfully after downloading the new module dependencies with `go mod tidy`.

## Surprises & Discoveries

- Observation: The repository already contains `postgres` configuration in `internal/domain/config.go` and a working `docker-compose.yml` with Postgres and pgAdmin, so the integration can be added without changing user-facing configuration shape.
  Evidence: `config.yml` already has a `postgres:` section and `docker-compose.yml` exposes `5432`.

- Observation: The current `App` lifecycle is the right place for database ownership because `cmd/run.go` already creates one application instance and shuts it down centrally.
  Evidence: `internal/app/app.go` owns startup and `App.Close` already coordinates component shutdown.

## Decision Log

- Decision: Use `pgx/v5` with `pgxpool` for connections and `squirrel` for SQL building instead of `database/sql` or an ORM.
  Rationale: This keeps the stack lightweight, Postgres-native, and parameterized by default while avoiding handwritten SQL string concatenation for common CRUD statements.
  Date/Author: 2026-05-15 / Codex

- Decision: Stop at the shared Postgres infrastructure repository and do not add concrete CRUD repositories yet.
  Rationale: The user explicitly requested only the Postgres repository for now. Future table-specific repositories can consume this infrastructure later.
  Date/Author: 2026-05-15 / Codex

## Outcomes & Retrospective

The repository now has a reusable PostgreSQL infrastructure layer and application lifecycle wiring, but no table-specific repositories yet. That matches the narrowed scope requested during implementation. The main lesson from validation was that `pgx` does not expose `sslmode` through the runtime parameter map the way a naive test might expect, so the unit test was updated to verify the observable effect for `sslmode=disable`: a nil TLS config.

## Context and Orientation

The program starts in `main.go`, which calls `cmd.Execute()`. The `run` command in `cmd/run.go` parses the YAML config, creates the logger, validates runtime settings, and calls `internal/app.New`. The `App` type in `internal/app/app.go` wires repositories into the use case and starts the HTTP server. The current HTTP endpoint is `GET /site-info` in `internal/delivery/http/handler.go`, which calls `Usecase.AnalyzeSite`.

The domain layer lives in `internal/domain`. It contains shared models such as `SiteInfo`, shared errors, and repository interfaces used by the use case layer. The use case layer lives in `internal/usecase` and should contain business logic rather than SQL or transport code.

The new database infrastructure will live under `internal/repository/postgres`. A "manager" here means a small component that owns the Postgres connection pool, validates configuration, provides a query builder configured for PostgreSQL placeholder syntax, exposes a health check with `Ping`, exposes transaction entry points, and closes the pool on shutdown. Future table-specific repositories can depend on this package instead of creating their own connections.

## Plan of Work

First, add Postgres dependencies to `go.mod`: `github.com/jackc/pgx/v5/pgxpool` for pooled connections and `github.com/Masterminds/squirrel` for safe query construction with `$1`, `$2`, and so on. Then add a new package `internal/repository/postgres` with a manager type that validates `domain.DatabaseConfig`, constructs a DSN, creates a `pgxpool.Pool`, exposes `Ping`, `Close`, `Pool`, `Builder`, and transaction helpers, and logs errors before returning them.

Next, update `internal/app/app.go` so `App.New` creates the Postgres manager from `cfg.Database`, pings the database during startup, and stores the manager on `App` for shutdown. `App.Close` must close the HTTP server, then the use case, then the database pool. The use case layer and HTTP behavior should remain unchanged because no concrete CRUD repositories are being introduced yet.

Finally, add unit tests for the manager config validation and transaction entry points. The tests should prove the DSN and defaults are assembled correctly and that the manager exposes the reusable primitives future repositories will need.

## Concrete Steps

Work from the repository root `D:\Projects\Golang projects\bookingBot`.

1. Add and edit the new Go files and migration files described in this plan.
2. Run formatting on modified Go files:

       gofmt -w internal\app\app.go internal\repository\postgres\manager.go internal\repository\postgres\manager_test.go

3. Build the project:

       go build ./...

4. Run the test suite:

       go test ./...

5. To observe the end-to-end flow manually after implementation, start PostgreSQL and the app:

       docker compose up -d postgres
       go run . run

   Startup should succeed only when PostgreSQL is reachable with the configured credentials.

## Validation and Acceptance

Acceptance is met when all of the following are true.

The project builds with `go build ./...` and the test suite passes with `go test ./...`.

`internal/app.New` creates and pings the shared Postgres manager during startup, and `App.Close` closes the pool during shutdown.

The manager exposes a PostgreSQL-configured `squirrel` builder and transaction entry points that future table-specific repositories can use without opening their own connections.

Running the service against a live Postgres instance succeeds when the database is reachable and fails fast during startup when the database is unavailable or misconfigured.

## Idempotence and Recovery

The Go code and migration file edits are additive and safe to re-run. Formatting, build, and test commands are idempotent.

If the database is unavailable, `internal/app.New` should fail during `Ping` and return an error instead of letting the service start partially initialized. This is safe to retry after PostgreSQL is running again.

## Artifacts and Notes

Expected startup behavior:

    app.New(...) -> postgres.New(...) -> dbManager.Ping(...)
    startup fails before serving HTTP if PostgreSQL is unreachable

Validation evidence:

    go build ./...   -> success
    go test ./...    -> success

## Interfaces and Dependencies

Add these dependencies in `go.mod`:

- `github.com/jackc/pgx/v5` and `github.com/jackc/pgx/v5/pgxpool` for Postgres connectivity.
- `github.com/Masterminds/squirrel` for safe query generation with PostgreSQL placeholder formatting.

At the end of the change, these interfaces and constructors must exist.

In `internal/repository/postgres/manager.go`, define:

    type QueryExecutor interface {
        Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
        Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
        QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
    }

    type Manager struct { ... }

    func New(cfg domain.DatabaseConfig, log zerolog.Logger) (*Manager, error)
    func (m *Manager) Ping(ctx context.Context) error
    func (m *Manager) Close()
    func (m *Manager) BeginTx(ctx context.Context) (pgx.Tx, error)
    func (m *Manager) Pool() *pgxpool.Pool
    func (m *Manager) Builder() squirrel.StatementBuilderType

Revision note: updated the scope after implementation started to remove table-specific repositories and keep only the shared Postgres infrastructure requested by the user.
Revision note: updated progress and retrospective after implementation and successful validation.
