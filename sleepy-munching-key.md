# Plan: Add Multi-Node Process Monitor (/monitor/process)

## Context

The current process monitoring (`/hosts/process/process`) only monitors a single node using WebSocket. The user wants a new "多机进程监控" (Multi-Node Process Monitor) page at `/monitor/process` that aggregates process data from all registered nodes via the core backend, following the same pattern used by `/monitor/containers` and `/monitor/dashboard`.

## Approach

Unlike the single-node WebSocket approach, the multi-node version will use HTTP polling (every 5s) to fetch process lists from each node through core's proxy. This avoids managing N WebSocket connections and follows the established monitor pattern.

---

## Changes

### 1. Backend: Add `GetNodeProcesses` to PanelClient

**File:** `core/utils/panel_client.go`

- Add a `GetProcesses()` method that calls `GET /process/list` on the agent (we'll also add this agent endpoint)
- Returns `[]PsProcessData` (re-use agent's struct definition or define equivalent in core)

### 2. Backend: Add process list endpoint to Agent

**File:** `agent/app/api/v2/process.go` — add `ListProcesses` handler

**File:** `agent/router/ro_process.go` — register `POST /process/list` route

**File:** `agent/app/service/process.go` — extract the process listing logic from the WebSocket handler into a reusable `ListProcesses(filter)` method that returns `[]PsProcessData`

This new HTTP endpoint accepts optional filters (pid, name, username) and returns JSON — same data as the WebSocket `ps` message but via REST.

### 3. Backend: Add `GetNodeProcesses` to Core Node API

**File:** `core/app/api/v2/node.go` — add `GetNodeProcesses(c *gin.Context)` handler
- Gets node ID from URL param
- Gets client via `nodeService.GetClient(id)`
- Calls `client.GetProcesses(filters)` to fetch from agent
- Returns process list with node info attached

**File:** `core/router/ro_node.go` — add route `GET /:id/processes`

### 4. Frontend: Add API function

**File:** `frontend/src/api/modules/node.ts`
- Add `getNodeProcesses(id: number)` → `GET /core/nodes/${id}/processes`

### 5. Frontend: Add route

**File:** `frontend/src/routers/modules/monitor.ts`
- Add child route: `path: '/monitor/process'`, name `'MonitorProcess'`
- Component: `import('@/views/monitor/process/index.vue')`

### 6. Frontend: Create the view

**File:** `frontend/src/views/monitor/process/index.vue` (new)

Based on `/monitor/containers/index.vue` pattern:
- Node selector dropdown at top
- Fetch process list for selected node via `getNodeProcesses(id)`
- Auto-refresh every 5 seconds
- Table showing: PID, Name, User, CPU%, Memory, Threads, Connections, Status, Start Time
- No stop/kill action (read-only monitoring across nodes)

### 7. Menu registration

**File:** `core/init/migration/helper/menu.go`
- Add child menu item under Monitor-Menu (ID "14"):
  - ID: `"145"`, Title: `"menu.monitorProcess"`, Path: `"/monitor/process"`, Label: `"MonitorProcess"`, Sort: 500

**File:** `core/constant/common.go`
- Add `"/monitor/process": {}` to `WebUrlMap`

### 8. i18n

**File:** `frontend/src/lang/modules/zh.ts` — add `monitorProcess: '多机进程监控'` under `menu`
**File:** `frontend/src/lang/modules/en.ts` — add `monitorProcess: 'Process Monitor'` under `menu`

---

## Verification

1. `cd frontend && pnpm run type-check` — TypeScript compiles
2. `cd core && go build ./...` — Core backend compiles
3. `cd agent && go build ./...` — Agent backend compiles
4. Manual test: start dev servers, navigate to `/monitor/process`, verify node selection and process data display
