# AGENTS.md

## Commands

```bash
# Run the server (requires SENTRY_ORG and SENTRY_AUTH_TOKEN env vars)
go run main.go serve

# Run tests
make test

# Build dev version
make build-dev

# Build staging version
make build-staging

# Build production (requires VERSION and CHANNEL env vars)
make build

# Lint
golangci-lint run
```

## Environment Variables

| Variable | Required | Default |
|----------|----------|---------|
| `SENTRY_ORG` | Yes | - |
| `SENTRY_AUTH_TOKEN` | Yes | - |
| `GLANCE_SENTRY_PORT` | No | 8099 |
| `GLANCE_SENTRY_HOST` | No | 127.0.0.1 |

## Build Notes

- All builds use vendor mode: `GOFLAGS="-mod=vendor"`
- Production builds require `VERSION` and `CHANNEL` env vars
- Build outputs go to `build/` directory

## Tech Stack

- Go 1.26.2
- Cobra (CLI)
- Uber FX (DI)
- golangci-lint v2

## Project Structure

- `main.go` - Entry point, creates root `cobra.Command`
- `cmd/serve.go` - HTTP server command, uses `*ServeCmd` struct pattern
- `app/kernel.go` - DI kernel definition (`fx.Module`)
- `app/bootstrap/` - Initialization: `initialization.go`, `module/` (logger, sentry)
- `app/config/env.go` - Env struct with `Load()` and `Validate()`
- `app/service/sentry/` - External client: `client.go`, `client_interface.go`
- `app/model/` - Data models: one file per domain (`release/entity.go`)
- `widget.yml` - Glance widget config

## Patterns Used

- **Interface + Implementation**: `service/x/*_interface.go` defines interface, `service/x/x.go` implements
- **Compile-time check**: `var _ InterfaceType = (*ImplementationType)(nil)`
- **fx modules**: `bootstrap/module/` wraps services with `AppXxx` or `XxxClient` types
- **Client pattern**: External service clients named `Client` in `service/sentry/` package
- **Env config**: Struct in `config/env.go` with typed fields, `Load()` method, validation helpers
- **Cobra commands**: `XxxCmd` struct with `NewXxxCmd()`, `Command()`, `run()`, `execute()` pattern

## Lint Configuration

- `.golangci.yml` uses v2 format
- Enabled checks: `staticcheck` with `-ST1005`, `-ST1000` exclusions