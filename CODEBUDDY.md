# CODEBUDDY.md

This file provides guidance to CodeBuddy Code when working with code in this repository.

## Project Overview

1Panel is an open-source VPS control panel with native AI agent support. The project is a Go + Vue.js monorepo with two backend services (core and agent) and one frontend SPA. The Go module path is `github.com/1Panel-dev/1Panel`.

## Repository Structure

```
├── core/           # 1panel-core: central management service (auth, settings, agent orchestration)
├── agent/          # 1panel-agent: node-level service (Docker, apps, websites, databases, firewall, etc.)
├── frontend/       # Vue 3 SPA served by the core service
├── ci/             # Release scripts (downloads 1pctl, install.sh, systemd units, GeoIP)
└── .github/        # CI workflows
```

## Dual-Backend Architecture

**core** is the control plane — it manages authentication (session + passkey + TOTP), agent registration, settings, groups, backup orchestration, script library, and serves the frontend static assets. It listens on a TCP port with optional TLS/mTLS via cmux (HTTP+HTTPS on same port).

**agent** is the worker node — it handles all node-level operations: Docker/container management, app marketplace installs, website/nginx config, databases (MySQL/PostgreSQL/Redis/MongoDB), firewall, fail2ban, cronjobs, file management, monitoring, SSL, snapshots, and AI (Ollama, MCP server, OpenClaw). It runs in two modes:
- **Master mode** (`global.IsMaster == true`): communicates over Unix socket at `/etc/1panel/agent.sock`
- **Remote node mode**: listens on a TCP port with mTLS (client certs required)

Both services share the same layered architecture within their `app/` directories:

```
app/
├── api/v2/       # HTTP handlers (Gin controllers) — request binding and response
├── dto/          # Data transfer objects (request/response structs)
│   ├── request/
│   └── response/
├── model/        # GORM database models
├── repo/         # Repository layer — database queries
├── service/      # Business logic
├── task/         # Async task runner
└── provider/     # (agent only) External providers like OpenClaw
```

Routing: `router/` defines Gin route groups. Core routes are under `/api/v2/core/*`, agent routes under `/api/v2/*`. Both use middleware in `middleware/` for auth, session, CSRF, logging.

Initialization sequence for each service is in `server/server.go` → `init/` packages (viper, db, migration, i18n, validator, cron, session/hook, etc.).

## Build Commands

```bash
# Full build (frontend + both backends for current Linux arch)
make build_all

# Individual targets
make build_frontend          # cd frontend && pnpm install && npm run build:pro
make build_core_on_linux     # CGO_ENABLED=0 go build -o build/1panel-core core/cmd/server/main.go
make build_agent_on_linux    # CGO_ENABLED=0 go build -o build/1panel-agent agent/cmd/server/main.go

# macOS cross-compilation (targets linux/amd64)
make build_on_local          # clean_assets + frontend + core + agent for darwin

# Compress binaries (requires upx)
make upx_bin
```

Release builds use GoReleaser (`.goreleaser.yaml`) with `-tags=xpack` build flag and target linux/{amd64,arm64,arm,ppc64le,s390x,riscv64}.

## Frontend

- **Stack**: Vue 3 + TypeScript + Vite + Element Plus + Pinia + Tailwind CSS
- **Package manager**: npm
- **Dev server**: `cd frontend && npm run dev` (port 4004, proxies `/api/v2` to `localhost:9999`)
- **Build**: `npm run build:pro` (output goes to `core/cmd/server/web/` — embedded into core binary via `go:embed`)
- **Lint**: `npm run lint:eslint` (ESLint), `npm run lint:prettier` (Prettier)
- **Type check**: `npm run type-check` (vue-tsc --noEmit)
- **i18n**: `src/lang/` — supports 10+ languages
- **Path alias**: `@` → `src/`
- **Auto-import**: Vue/Vue-Router APIs and Element Plus components are auto-imported via unplugin

Key frontend directories:
- `src/views/` — page components (ai, app-store, container, cronjob, database, host, log, login, setting, share, terminal, toolbox, website)
- `src/api/` — Axios API calls
- `src/store/` — Pinia stores
- `src/routers/` — Vue Router config
- `src/composables/` — Vue composables

## Backend Conventions

- **Database**: SQLite (via glebarez/sqlite driver + GORM). Migrations in `init/migration/`.
- **CLI**: Cobra-based. Core has subcommands (reset, restore, update, user-info, version, listen-ip); agent has only root.
- **Config**: Viper reads YAML config. Global config at `global.CONF`.
- **Logging**: Logrus (`global.LOG`).
- **i18n**: go-i18n with locale files in `i18n/`.
- **API docs**: Swag/Swagger annotations on main.go. Access at `/{entrance}/1panel/swagger/`.
- **Error handling**: Custom business errors in `buserr/`.
- **Build tag**: `xpack` — pro/enterprise features are gated behind this build tag. Files like `entry_xpack.go` use `//go:build xpack`.

## API Authentication

- Core uses session-based auth with CSRF protection. The API key header format is `1Panel-Token: md5('1panel' + API-Key + UnixTimestamp)` with `1Panel-Timestamp` header.
- Agent (remote mode) uses mTLS with client certificate verification.

## CI

- **Release**: Triggered on `v*` tags → builds frontend, then GoReleaser for both binaries, uploads to GitHub Releases (draft), Alibaba OSS, and Cloudflare R2.
- **Other workflows**: SonarCloud scan, LLM code review, PR labeler, issue translator, typos check, Gitee sync.
