package dto

import "time"

// NodeCreate 创建节点请求
type NodeCreate struct {
	Name                 string `json:"name" validate:"required"`
	Host                 string `json:"host" validate:"required"`
	Port                 int    `json:"port" validate:"required,min=1,max=65535"`
	APIKey               string `json:"apiKey" validate:"required"`
	Tags                 string `json:"tags"`
	ServerName           string `json:"serverName"`
	Security             string `json:"security"` // "none" | "tls"
	AllowInsecure        bool   `json:"allowInsecure"`
	PinnedPeerCertSHA256 string `json:"pinnedPeerCertSha256"`
	CertPem              string `json:"certPem"` // PEM certificate text; auto-computes PinnedPeerCertSHA256 if provided
}

// NodeTest 测试节点连接请求（编辑时 APIKey 可选，通过 id 回填）
type NodeTest struct {
	ID                   uint   `json:"id"`
	Name                 string `json:"name" validate:"required"`
	Host                 string `json:"host" validate:"required"`
	Port                 int    `json:"port" validate:"required,min=1,max=65535"`
	APIKey               string `json:"apiKey"`
	Tags                 string `json:"tags"`
	ServerName           string `json:"serverName"`
	Security             string `json:"security"`
	AllowInsecure        bool   `json:"allowInsecure"`
	PinnedPeerCertSHA256 string `json:"pinnedPeerCertSha256"`
	CertPem              string `json:"certPem"`
}

// NodeUpdate 更新节点请求
type NodeUpdate struct {
	Name                 string `json:"name" validate:"required"`
	Host                 string `json:"host" validate:"required"`
	Port                 int    `json:"port" validate:"required,min=1,max=65535"`
	APIKey               string `json:"apiKey"`
	Tags                 string `json:"tags"`
	ServerName           string `json:"serverName"`
	Security             string `json:"security"`
	AllowInsecure        bool   `json:"allowInsecure"`
	PinnedPeerCertSHA256 string `json:"pinnedPeerCertSha256"`
	CertPem              string `json:"certPem"`
}

// NodeInfo 节点信息响应
type NodeInfo struct {
	ID                   uint       `json:"id"`
	Name                 string     `json:"name"`
	Host                 string     `json:"host"`
	Port                 int        `json:"port"`
	Status               int        `json:"status"`
	LastSeen             *time.Time `json:"lastSeen"`
	Tags                 string     `json:"tags"`
	IsDefault            bool       `json:"isDefault"`
	ServerName           string     `json:"serverName"`
	Security             string     `json:"security"`
	AllowInsecure        bool       `json:"allowInsecure"`
	PinnedPeerCertSHA256 string     `json:"pinnedPeerCertSha256"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

// MonitorSearchReq 监控历史查询请求
type MonitorSearchReq struct {
	NodeID     uint   `json:"nodeId" validate:"required"`
	MetricType string `json:"metricType" validate:"required"`
	StartTime  string `json:"startTime" validate:"required"`
	EndTime    string `json:"endTime" validate:"required"`
}

// MonitorHistoryItem 监控历史数据点
type MonitorHistoryItem struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

// NodeDashboard 节点仪表盘数据
type NodeDashboard struct {
	NodeID    uint    `json:"nodeId"`
	NodeName  string  `json:"nodeName"`
	Status    int     `json:"status"`
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
	Disk      float64 `json:"disk"`
	NetUp     float64 `json:"netUp"`
	NetDown   float64 `json:"netDown"`
	IORead    float64 `json:"ioRead"`
	IOWrite   float64 `json:"ioWrite"`
	Load1     float64 `json:"load1"`
	Load5     float64 `json:"load5"`
	Load15    float64 `json:"load15"`
	Hostname  string  `json:"hostname"`
	OSVersion string  `json:"osVersion"`
	Uptime    uint64  `json:"uptime"`

	MemoryTotal int64 `json:"memoryTotal"`
	MemoryUsed  int64 `json:"memoryUsed"`
	MemoryAvail int64 `json:"memoryAvail"`

	SwapTotal int64 `json:"swapTotal"`
	SwapUsed  int64 `json:"swapUsed"`

	DiskTotal int64 `json:"diskTotal"`
	DiskUsed  int64 `json:"diskUsed"`

	IPAddress string `json:"ipAddress"`
}

// NodeHeatmapReq 节点热力图请求
type NodeHeatmapReq struct {
	Hours       int `form:"hours" json:"hours"`             // 时间范围（小时），默认24
	StepMinutes int `form:"stepMinutes" json:"stepMinutes"` // 时间粒度（分钟），默认5
}

// NodeHeatmapData 节点热力图数据
type NodeHeatmapData struct {
	Times  []string `json:"times"`  // 时间轴，如 ["01-02 15:00", "01-02 15:05", ...]
	Nodes  []string `json:"nodes"`  // 节点名称列表
	Values [][]int  `json:"values"` // 二维数组，values[nodeIndex][timeIndex] = 0/1/2
}
