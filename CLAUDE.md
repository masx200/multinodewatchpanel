# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## 项目概述

MultiNodeWatchPanel 是一个基于 1Panel v2 架构的多节点服务器监控管理面板。采用
Go + Vue.js 单体仓库结构，包含两个后端服务（core 和 agent）和一个前端 SPA。Go
模块路径为 `github.com/masx200/multinodewatchpanel/core` 和
`github.com/masx200/multinodewatchpanel/agent`。

## 构建命令

```bash
# 完整构建（前端 + 两个后端）
make build_all

# 单独构建
make build_frontend          # cd frontend && pnpm install && npm run build:pro
make build_core_on_linux     # CGO_ENABLED=0 go build -o build/1panel-core core/cmd/server/main.go
make build_agent_on_linux    # CGO_ENABLED=0 go build -o build/1panel-agent agent/cmd/server/main.go

# macOS 交叉编译（目标 linux/amd64）
make build_on_local

# 压缩二进制文件（需要 upx）
make upx_bin
```

### 前端命令

```bash
cd frontend
pnpm install              # 安装依赖
pnpm dev                  # 开发服务器（端口 4004，代理 /api/v2 -> localhost:9999）
pnpm run build:pro        # 生产构建（输出到 ../core/cmd/server/web/，嵌入 Go 二进制）
pnpm run type-check       # TypeScript 类型检查（vue-tsc --noEmit）
pnpm run lint:eslint      # ESLint 检查
pnpm run lint:prettier    # Prettier 格式化
```

### Go 测试

```bash
# 运行单个包的测试
cd core && go test ./utils/cmd/...
cd agent && go test ./utils/...
```

## 双后端架构

### core（控制面板）

core 是控制平面，负责：认证（session + passkey + TOTP）、agent
注册、设置管理、分组、备份编排、脚本库、前端静态资源服务。

- 入口：`core/cmd/server/main.go`
- 监听 TCP 端口，支持通过 cmux 在同一端口上同时提供 HTTP + HTTPS（可选
  TLS/mTLS）
- CLI 基于
  Cobra，子命令包括：reset、restore、update、user-info、version、listen-ip
- 路由前缀：`/api/v2/core/*`

### agent（工作节点）

agent 是工作节点，负责所有节点级操作：Docker/容器管理、应用市场安装、网站/Nginx
配置、数据库（MySQL/PostgreSQL/Redis/MongoDB）、防火墙、fail2ban、定时任务、文件管理、监控、SSL、快照、AI（Ollama、MCP
server、OpenClaw）。

- 入口：`agent/cmd/server/main.go`
- **Master 模式**：通过 Unix socket `/etc/1panel/agent.sock` 与 core 通信
- **远程节点模式**：监听 TCP 端口，使用 mTLS（客户端证书验证）
- 路由前缀：`/api/v2/*`

### 共享分层架构（core 和 agent 的 `app/` 目录结构相同）

```
app/
├── api/v2/       # HTTP 处理器（Gin 控制器）— 请求绑定和响应
├── dto/          # 数据传输对象
│   ├── request/  # 请求结构体
│   └── response/ # 响应结构体
├── model/        # GORM 数据库模型
├── repo/         # 仓库层 — 数据库查询
├── service/      # 业务逻辑
├── task/         # 异步任务运行器
└── provider/     # 外部提供者（仅 agent）
```

### 其他后端目录

- `router/` — Gin 路由组定义
- `middleware/` — 认证、session、CSRF、日志中间件
- `init/` —
  初始化序列（viper、db、migration、i18n、validator、cron、session/hook 等）
- `server/server.go` — 服务启动入口
- `buserr/` — 自定义业务错误
- `i18n/` — go-i18n 本地化文件
- `global/` — 全局配置（`global.CONF`）和日志（`global.LOG`）

## 前端架构

详细的前端架构见 `frontend/CLAUDE.md`。

- **技术栈**：Vue 3 (Composition API + `<script setup>`) + TypeScript + Vite 7 +
  Element Plus + Pinia + Tailwind CSS 4
- **包管理器**：pnpm
- **路径别名**：`@` → `src/`
- **自动导入**：Vue/vue-router API 和 Element Plus 组件自动导入
- **构建输出**：嵌入 core Go 二进制（`go:embed`）
- **i18n**：`src/lang/` — 支持 10+ 种语言
- **多节点**：`GlobalStore.currentNode` 追踪当前节点（`local` 为
  master），`isMaster` getter 检查是否在 master 节点

## 后端约定

- **数据库**：SQLite（通过 glebarez/sqlite 驱动 + GORM），迁移在
  `init/migration/`
- **配置**：Viper 读取 YAML，全局配置 `global.CONF`
- **日志**：Logrus（`global.LOG`）
- **API 文档**：Swag/Swagger 注解，访问路径 `/{entrance}/1panel/swagger/`
- **构建标签**：`xpack` — Pro/企业功能通过此构建标签控制，文件如
  `entry_xpack.go` 使用 `//go:build xpack`

## API 认证

- Core 使用基于 session 的认证 + CSRF 保护。API Key
  格式：`1Panel-Token: md5('1panel' + API-Key + UnixTimestamp)`，附带
  `1Panel-Timestamp` 头
- Agent（远程模式）使用 mTLS 客户端证书验证

## Docker 部署

- `docker-compose.yml` 配置了完整的服务，端口 9999
- `Dockerfile` 基于 Debian slim，安装 docker-cli、docker-compose 等依赖
- 环境变量控制面板配置（端口、用户名、密码、安全入口等）

## CI/CD

- 发布流程：`v*` 标签触发 → 构建前端 → GoReleaser 构建两个二进制 → 上传到 GitHub
  Releases
- GoReleaser 配置：`.goreleaser.yaml`，使用 `-tags=xpack` 构建标签，目标平台
  linux/{amd64,arm64,arm,ppc64le,s390x,riscv64}
- 其他工作流：SonarCloud 扫描、LLM 代码审查、PR 标签、Issue 翻译、拼写检查
