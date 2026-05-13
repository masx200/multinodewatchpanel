# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Project Overview

This is the frontend for **MultiNodeWatchPanel**, a multi-node server monitoring
and management panel built on top of the 1Panel v2 architecture. It is a Vue 3
SPA that connects to a Go backend via `/api/v2` proxy.

## Commands

```bash
# Install dependencies (uses pnpm)
pnpm install

# Dev server (port 4004, proxies /api/v2 -> localhost:9999)
pnpm dev

# Type checking
pnpm run type-check

# Build for production (output -> ../core/cmd/server/web)
pnpm run build:pro

# Lint
pnpm run lint:eslint
pnpm run lint:prettier
```

## Architecture

**Tech stack:** Vue 3 (Composition API + `<script setup>`), TypeScript
(non-strict mode), Vite 7, Element Plus, Pinia, vue-router, vue-i18n, Tailwind
CSS 4, CodeMirror 6, Monaco Editor, xterm.js, ECharts.

**Path alias:** `@` maps to `src/`.

**Auto-imports:** Vue and vue-router APIs are auto-imported via
`unplugin-auto-import`. Element Plus components are auto-resolved via
`unplugin-vue-components`. Both generate `auto-imports.d.ts` and
`components.d.ts`.

**Build output:** Production builds go to `../core/cmd/server/web` — this
frontend is embedded into the Go binary.

### Key Directory Structure

- `src/routers/modules/` — Each file is a top-level menu section (monitor,
  container, database, etc.). Routes use `sort` field for ordering.
  `meta.requiresAuth: false` skips auth guard.
- `src/api/modules/` — Axios-based API modules. One file per domain (app.ts,
  container.ts, node.ts, etc.). All API calls go through `/api/v2` proxy.
- `src/store/modules/` — Pinia stores with persistence
  (`pinia-plugin-persistedstate`). Key stores: `global` (auth state, theme,
  current node), `menu`, `tabs`, `terminal`, `process`.
- `src/views/` — Page-level components organized by feature area. Each area has
  its own directory.
- `src/components/` — Shared reusable components registered globally via
  `src/components/index.ts` plugin.
- `src/lang/modules/` — i18n locale files (zh, en, ja, ko, ru, etc.).
  Lazy-loaded per locale.
- `src/extensions/` — Plugin/extension system (routes, theme, xpack features).
- `src/styles/var.scss` — Global SCSS variables, auto-injected into every `.vue`
  file via Vite `additionalData`.

### Routing & Auth

- Router guard in `src/routers/index.ts` checks `GlobalStore.isLogin` for auth.
- The `entrance` route pattern (`/:code?`) handles the security entrance URL.
- Active menu tab caching uses `localStorage` with keys like
  `cachedRoute{activeMenu}`.

### State Management

- `GlobalStore` holds auth state, theme config, current node (`currentNode`),
  and multi-node awareness (`isMaster`, `masterAlias`).
- All stores persist to localStorage automatically.

### Multi-Node Architecture

This is a multi-node monitoring panel. The `currentNode` in GlobalStore tracks
whether the user is on `local` (master) or a remote node. Components use
`node-select` for node switching. The `isMaster` getter checks
`currentNode === 'local'`.

## Code Style

- **Prettier:** 4-space indent, single quotes, trailing commas, 120 char line
  width.
- **ESLint:** vue3-recommended + @typescript-eslint/recommended + prettier
  integration. `no-explicit-any` is off.
- **SCSS:** Global variables in `src/styles/var.scss` are available in all
  components without explicit import.
- **Components:** Use Element Plus components directly (auto-resolved). Custom
  global components are registered in `src/components/index.ts`.
