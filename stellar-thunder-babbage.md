# MultiNodeWatchPanel - 多机监控面板实现计划

## 项目概述

基于 1Panel 开源框架开发的多机监控面板，实现对多台 1Panel 服务器的统一监控管理。

---

## 一、技术架构

### 1.1 技术选型
| 层级 | 技术 | 说明 |
|------|------|------|
| 前端 | Vue 3 + TypeScript + Vite | 复用 1Panel 前端 |
| UI | Element Plus + Tailwind CSS | 复用 1Panel UI |
| 后端 | Go + Gin + GORM | 复用 1Panel 后端 |
| 数据库 | SQLite (开发) / MySQL (生产) | 复用 1Panel 数据库 |
| 监控协议 | HTTP REST API | 调用 1Panel API |

### 1.2 项目结构
```
multinode-watchpanel/
├── frontend/           # 前端 (1Panel frontend + 自定义视图)
│   └── src/
│       ├── views/
│       │   ├── monitor/           # 新增：监控视图
│       │   │   ├── dashboard/     # 仪表盘
│       │   │   ├── hosts/         # 主机列表
│       │   │   ├── containers/     # 容器监控
│       │   │   ├── monitor/       # 历史监控
│       │   │   └── gpu/           # GPU监控
│       │   └── setting/
│       │       └── nodes/         # 节点管理 (新增)
│       ├── api/modules/
│       │   ├── node.ts            # 节点管理API
│       │   └── monitor.ts         # 监控数据API
│       └── store/modules/
│           └── node.ts            # 节点状态管理
├── core/               # 后端 (1Panel core + 自定义模块)
│   └── app/
│       └── model/
│           ├── node.go             # 节点配置模型
│           └── monitor_data.go     # 历史数据模型
│       └── router/
│           └── node_router.go      # 节点路由
│       └── service/
│           ├── node_service.go     # 节点管理服务
│           └── collector.go        # 数据采集服务
│       └── utils/
│           └── panel_client.go     # 1Panel API客户端
└── docs/
    └── API_INTEGRATION.md         # 1Panel API集成文档
```

---

## 二、数据库设计

### 2.1 节点配置表 (host_nodes)
```sql
CREATE TABLE host_nodes (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    name         VARCHAR(64)  NOT NULL COMMENT '节点名称',
    host         VARCHAR(255) NOT NULL COMMENT '1Panel地址',
    port         INT          NOT NULL DEFAULT 9999 COMMENT '端口',
    api_key      VARCHAR(128) NOT NULL COMMENT 'API密钥(加密存储)',
    api_secret   VARCHAR(128)          COMMENT 'API密钥原始值(仅写入时)',
    status       TINYINT      DEFAULT 0 COMMENT '状态: 0=离线 1=在线 2=连接失败',
    last_seen    DATETIME              COMMENT '最后在线时间',
    tags         VARCHAR(255)          COMMENT '标签(逗号分隔)',
    is_default   TINYINT      DEFAULT 0 COMMENT '是否默认节点',
    created_at   DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 2.2 监控历史数据表 (monitor_history)
```sql
CREATE TABLE monitor_history (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    node_id      BIGINT       NOT NULL COMMENT '节点ID',
    metric_type  VARCHAR(32)  NOT NULL COMMENT '指标类型: cpu/memory/disk/io/network',
    metric_value DECIMAL(10,2) NOT NULL COMMENT '指标值',
    recorded_at  DATETIME     NOT NULL COMMENT '记录时间',
    created_at   DATETIME     DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_node_time (node_id, recorded_at),
    INDEX idx_type_time (metric_type, recorded_at)
);
```

### 2.3 系统配置表 (system_configs)
```sql
CREATE TABLE system_configs (
    key   VARCHAR(64) PRIMARY KEY,
    value TEXT,
    description VARCHAR(255)
);

-- 初始配置
INSERT INTO system_configs (key, value, description) VALUES
('data_retention_days', '30', '监控数据保留天数'),
('collect_interval', '60', '数据采集间隔(秒)'),
('dashboard_refresh', '10', '仪表盘刷新间隔(秒)');
```

---

## 三、后端模块设计

### 3.1 1Panel API 客户端 (utils/panel_client.go)

```go
type PanelClient struct {
    BaseURL    string
    APIKey     string
    HTTPClient *http.Client
}

func NewPanelClient(host string, port int, apiKey string) *PanelClient

// 认证头生成
func (c *PanelClient) AuthHeaders() (map[string]string, error)

// 核心监控接口
func (c *PanelClient) GetDashboardOS() (*OsInfo, error)
func (c *PanelClient) GetDashboardBase(ioOption, netOption string) (*DashboardBase, error)
func (c *PanelClient) GetDashboardCurrent(ioOption, netOption string) (*DashboardCurrent, error)
func (c *PanelClient) GetTopCPU() ([]Process, error)
func (c *PanelClient) GetTopMemory() ([]Process, error)

// 容器监控
func (c *PanelClient) GetContainerStatus() (*ContainerStatus, error)
func (c *PanelClient) ListContainers(req ListContainerReq) ([]ContainerInfo, error)
func (c *PanelClient) GetContainerStats(containerID string) (*ContainerStats, error)

// 历史监控
func (c *PanelClient) SearchMonitor(req MonitorSearchReq) ([]MonitorData, error)
```

### 3.2 节点管理服务 (service/node_service.go)

```go
type NodeService struct {
    db          *gorm.DB
    clients     sync.Map  // map[uint64]*PanelClient
    collector   *CollectorService
}

func (s *NodeService) CreateNode(req CreateNodeReq) error
func (s *NodeService) UpdateNode(id uint64, req UpdateNodeReq) error
func (s *NodeService) DeleteNode(id uint64) error
func (s *NodeService) ListNodes(req ListNodeReq) ([]HostNode, error)
func (s *NodeService) TestConnection(req TestNodeReq) (*TestResult, error)
func (s *NodeService) GetNodeStatus(id uint64) (string, error)
func (s *NodeService) GetClient(nodeID uint64) (*PanelClient, error)
```

### 3.3 数据采集服务 (service/collector.go)

```go
type CollectorService struct {
    db          *gorm.DB
    nodes       map[uint64]*PanelClient
    interval    time.Duration
    retention   int  // 保留天数
    stopChan    chan struct{}
}

func (c *CollectorService) Start() error
func (c *CollectorService) Stop()
func (c *CollectorService) CollectNode(nodeID uint64) error
func (c *PanelClient) CleanupOldData(retentionDays int) error
```

### 3.4 监控数据 API

```go
// GET /api/v2/nodes/dashboard/:id
func GetNodeDashboard(c *gin.Context)

// GET /api/v2/nodes/containers/:id
func GetNodeContainers(c *gin.Context)

// GET /api/v2/nodes/containers/:id/stats
func GetContainerStats(c *gin.Context)

// POST /api/v2/nodes/monitor/search
func SearchMonitorHistory(c *gin.Context)

// GET /api/v2/nodes/monitor/current/:id
func GetCurrentMonitor(c *gin.Context)

// GET /api/v2/nodes/top/:id
func GetTopProcesses(c *gin.Context)
```

---

## 四、前端模块设计

### 4.1 路由配置 (routers/modules/host.ts)

```typescript
// 新增多机监控路由
const monitorRouter = {
    sort: 2,
    path: '/monitor',
    name: 'Monitor',
    component: Layout,
    meta: { title: 'menu.monitor', icon: 'p-monitor' },
    children: [
        {
            path: 'dashboard',
            name: 'MonitorDashboard',
            component: () => import('@/views/monitor/dashboard/index.vue'),
            meta: { title: 'menu.monitorDashboard' },
        },
        {
            path: 'hosts',
            name: 'MonitorHosts',
            component: () => import('@/views/monitor/hosts/index.vue'),
            meta: { title: 'menu.monitorHosts' },
        },
        {
            path: 'containers',
            name: 'MonitorContainers',
            component: () => import('@/views/monitor/containers/index.vue'),
            meta: { title: 'menu.monitorContainers' },
        },
        {
            path: 'history',
            name: 'MonitorHistory',
            component: () => import('@/views/monitor/monitor/index.vue'),
            meta: { title: 'menu.monitorHistory' },
        },
    ],
};

// 节点管理路由
const nodeSettingRouter = {
    path: '/settings/nodes',
    name: 'NodeSetting',
    component: () => import('@/views/setting/nodes/index.vue'),
    meta: { title: 'menu.nodeSetting', icon: 'p-host' },
};
```

### 4.2 节点管理页面 (views/setting/nodes/index.vue)

```vue
<template>
    <LayoutContent title="节点管理" :divider="true">
        <!-- 工具栏 -->
        <template #toolbar>
            <el-button type="primary" @click="openDialog('create')">
                {{ $t('commons.button.add') }}
            </el-button>
            <el-button @click="openSyncDialog">
                {{ $t('monitor.syncNodes') }}
            </el-button>
        </template>

        <!-- 节点列表 -->
        <el-table :data="tableData" v-loading="loading">
            <el-table-column prop="name" label="节点名称" />
            <el-table-column prop="host" label="地址" />
            <el-table-column prop="port" label="端口" width="80" />
            <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                    <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                        {{ row.status === 1 ? '在线' : '离线' }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column prop="lastSeen" label="最后在线" />
            <el-table-column label="操作" width="150">
                <template #default="{ row }">
                    <el-button link @click="openDialog('edit', row)">编辑</el-button>
                    <el-button link type="danger" @click="deleteNode(row)">删除</el-button>
                </template>
            </el-table-column>
        </el-table>

        <!-- 节点编辑对话框 -->
        <DialogPro v-model="dialogVisible" :title="dialogTitle">
            <el-form :model="nodeForm" :rules="rules" ref="formRef">
                <el-form-item label="节点名称" prop="name">
                    <el-input v-model="nodeForm.name" />
                </el-form-item>
                <el-form-item label="1Panel地址" prop="host">
                    <el-input v-model="nodeForm.host" placeholder="https://example.com" />
                </el-form-item>
                <el-form-item label="端口" prop="port">
                    <el-input-number v-model="nodeForm.port" :min="1" :max="65535" />
                </el-form-item>
                <el-form-item label="API密钥" prop="apiKey">
                    <el-input v-model="nodeForm.apiKey" type="password" show-password />
                </el-form-item>
                <el-form-item label="标签">
                    <el-select v-model="nodeForm.tags" multiple filterable allow-create>
                        <el-option label="生产环境" value="production" />
                        <el-option label="测试环境" value="test" />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="dialogVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submitForm" :loading="btnLoading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </DialogPro>
    </LayoutContent>
</template>
```

### 4.3 监控仪表盘 (views/monitor/dashboard/index.vue)

```vue
<template>
    <div class="dashboard-container">
        <!-- 节点选择器 -->
        <el-select v-model="currentNode" @change="loadDashboard">
            <el-option v-for="node in nodes" :key="node.id" :label="node.name" :value="node.id" />
        </el-select>

        <!-- 主机概览卡片 -->
        <el-row :gutter="20" class="info-cards">
            <el-col :span="6">
                <InfoCard icon="cpu" :title="'CPU: ' + stats.cpu" :value="stats.cpuPercent + '%'" />
            </el-col>
            <el-col :span="6">
                <InfoCard icon="memory" :title="'内存: ' + stats.memory" :value="stats.memPercent + '%'" />
            </el-col>
            <el-col :span="6">
                <InfoCard icon="disk" :title="'磁盘: ' + stats.disk" :value="stats.diskPercent + '%'" />
            </el-col>
            <el-col :span="6">
                <InfoCard icon="network" :title="'网络: ' + stats.network" :value="stats.netSpeed" />
            </el-col>
        </el-row>

        <!-- 实时图表 -->
        <el-row :gutter="20">
            <el-col :span="12">
                <ChartCard title="CPU 使用率">
                    <LineChart :data="cpuHistory" height="300px" />
                </ChartCard>
            </el-col>
            <el-col :span="12">
                <ChartCard title="内存使用率">
                    <LineChart :data="memHistory" height="300px" />
                </ChartCard>
            </el-col>
        </el-row>

        <!-- 容器状态 -->
        <el-row :gutter="20">
            <el-col :span="12">
                <ChartCard title="容器状态">
                    <PieChart :data="containerStats" height="250px" />
                </ChartCard>
            </el-col>
            <el-col :span="12">
                <ChartCard title="Top 进程">
                    <el-table :data="topProcesses" size="small">
                        <el-table-column prop="name" label="进程名" />
                        <el-table-column prop="cpu" label="CPU%" width="80" />
                        <el-table-column prop="mem" label="内存%" width="80" />
                    </el-table>
                </ChartCard>
            </el-col>
        </el-row>
    </div>
</template>
```

---

## 五、功能清单

### 5.1 核心功能

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 节点管理 | P0 | 添加/编辑/删除监控节点 |
| 节点连接测试 | P0 | 验证API连接 |
| 仪表盘展示 | P0 | 实时CPU/内存/磁盘/网络 |
| 容器监控 | P0 | 容器列表与状态 |
| 历史数据 | P1 | 监控历史曲线图 |
| 数据采集 | P0 | 后台定时采集 |
| 数据清理 | P1 | 自动清理过期数据 |
| 多节点聚合 | P2 | 统一视图 |

### 5.2 设置功能

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 数据保留时间 | P0 | 设置监控数据保留天数 |
| 采集间隔 | P1 | 设置数据采集频率 |
| 刷新间隔 | P1 | 仪表盘自动刷新 |
| 节点标签 | P2 | 分组管理 |

---

## 六、实现步骤

### Phase 1: 项目初始化
1. 复制 1Panel 前端代码到 `frontend/`
2. 复制 1Panel 后端代码到 `core/`
3. 初始化 Git 仓库

### Phase 2: 数据库扩展
1. 创建 `host_nodes` 表
2. 创建 `monitor_history` 表
3. 创建 `system_configs` 表
4. 编写数据库迁移脚本

### Phase 3: 后端核心模块
1. 实现 `PanelClient` (1Panel API 客户端)
2. 实现 `NodeService` (节点管理)
3. 实现 `CollectorService` (数据采集)
4. 注册路由并测试

### Phase 4: 前端核心模块
1. 添加节点管理页面
2. 添加监控仪表盘页面
3. 添加容器监控页面
4. 添加设置页面

### Phase 5: 完善与优化
1. 历史数据图表
2. 告警功能
3. 性能优化
4. 文档编写

---

## 七、关键文件清单

### 后端新增/修改
| 文件路径 | 说明 |
|----------|------|
| `core/app/model/node.go` | 节点模型 |
| `core/app/model/monitor_history.go` | 历史数据模型 |
| `core/app/model/system_config.go` | 系统配置模型 |
| `core/app/service/node_service.go` | 节点管理服务 |
| `core/app/service/collector_service.go` | 数据采集服务 |
| `core/app/utils/panel_client.go` | 1Panel API客户端 |
| `core/app/router/node_router.go` | 节点路由 |

### 前端新增/修改
| 文件路径 | 说明 |
|----------|------|
| `frontend/src/views/setting/nodes/index.vue` | 节点管理页面 |
| `frontend/src/views/monitor/dashboard/index.vue` | 监控仪表盘 |
| `frontend/src/views/monitor/hosts/index.vue` | 主机列表 |
| `frontend/src/views/monitor/containers/index.vue` | 容器监控 |
| `frontend/src/views/monitor/monitor/index.vue` | 历史监控 |
| `frontend/src/api/modules/node.ts` | 节点API |
| `frontend/src/api/modules/monitor.ts` | 监控API |
| `frontend/src/store/modules/node.ts` | 节点状态 |
| `frontend/src/routers/modules/monitor.ts` | 监控路由 |
| `frontend/src/lang/modules/*.ts` | 国际化文本 |

---

## 八、1Panel API 认证流程

```go
// Token 生成算法
func GenerateToken(apiKey string) (string, string) {
    timestamp := fmt.Sprintf("%d", time.Now().Unix())
    token := md5.Sum([]byte("1panel" + apiKey + timestamp))
    tokenHex := hex.EncodeToString(token[:])
    return tokenHex, timestamp
}

// 请求头设置
headers := map[string]string{
    "1Panel-Token":      token,
    "1Panel-Timestamp":  timestamp,
    "Content-Type":      "application/json",
}
```

---

## 九、部署说明

### 9.1 开发环境
```bash
# 启动后端
cd core
go build -o multinode-watchpanel ./cmd/server
./multinode-watchpanel

# 启动前端
cd frontend
npm install
npm run dev
```

### 9.2 生产环境
- 后端: 编译后部署，监听端口 9999
- 前端: `npm run build`，输出到 `core/cmd/server/web`
- 数据库: MySQL 5.7+ 或 SQLite

---

## 十、注意事项

1. **API 密钥安全**: 节点 API 密钥需加密存储
2. **时间同步**: 确保多节点服务器时间同步
3. **性能考虑**: 大量节点时需优化采集频率
4. **数据量控制**: 合理设置保留天数，避免数据库过大
