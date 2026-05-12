package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// PanelClientOptions 创建客户端的可选参数
type PanelClientOptions struct {
	// Security: "none" | "tls"
	Security            string
	// AllowInsecure 跳过证书校验（仅 tls 时有效）
	AllowInsecure       bool
	// PinnedPeerCertSHA256 pin 证书 SHA256（hex，大小写不敏感），空则不校验
	PinnedPeerCertSHA256 string
	// ServerName TLS SNI（覆盖 Host）
	ServerName          string
}

// PanelClient 1Panel Agent API 客户端
// BaseURL 应指向 agent 的 API 根路径，例如 http://host:port/api/v2
type PanelClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewPanelClient 创建新的 PanelClient（默认 http，无 TLS）
func NewPanelClient(host string, port int, apiKey string) *PanelClient {
	return NewPanelClientWithOptions(host, port, apiKey, PanelClientOptions{Security: "none"})
}

// NewPanelClientWithOptions 创建支持 TLS 选项的 PanelClient
func NewPanelClientWithOptions(host string, port int, apiKey string, opts PanelClientOptions) *PanelClient {
	scheme := "http"
	if strings.EqualFold(opts.Security, "tls") {
		scheme = "https"
	}
	// BaseURL 包含 /api/v2 前缀，agent 端 basePath 为 /api/v2
	baseURL := fmt.Sprintf("%s://%s:%d/api/v2", scheme, host, port)

	transport := buildTransport(opts)
	return &PanelClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		},
	}
}

// buildTransport 根据选项构建 HTTP Transport
func buildTransport(opts PanelClientOptions) http.RoundTripper {
	if !strings.EqualFold(opts.Security, "tls") {
		return http.DefaultTransport
	}

	tlsConfig := &tls.Config{}
	if opts.ServerName != "" {
		tlsConfig.ServerName = opts.ServerName
	}

	pinnedHash := strings.ToLower(strings.TrimSpace(opts.PinnedPeerCertSHA256))

	if opts.AllowInsecure && pinnedHash == "" {
		// 完全跳过证书校验
		tlsConfig.InsecureSkipVerify = true
	} else if pinnedHash != "" {
		// 自定义 VerifyPeerCertificate：校验叶证书 SHA256
		tlsConfig.InsecureSkipVerify = true // 跳过系统链校验，由 VerifyPeerCertificate 接管
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return fmt.Errorf("no peer certificate presented")
			}
			sum := sha256.Sum256(rawCerts[0])
			got := hex.EncodeToString(sum[:])
			if got != pinnedHash {
				return fmt.Errorf("certificate SHA256 mismatch: got %s, want %s", got, pinnedHash)
			}
			return nil
		}
	}

	return &http.Transport{
		TLSClientConfig: tlsConfig,
	}
}

// GenerateToken 生成 1Panel API 认证 Token
func GenerateToken(apiKey string) (token, timestamp string) {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	h := md5.Sum([]byte("1panel" + apiKey + ts))
	return hex.EncodeToString(h[:]), ts
}

// AuthHeaders 返回认证请求头
func (c *PanelClient) AuthHeaders() map[string]string {
	token, ts := GenerateToken(c.APIKey)
	return map[string]string{
		"1Panel-Token":     token,
		"1Panel-Timestamp": ts,
		"Content-Type":     "application/json",
	}
}

// doRequest 执行 HTTP 请求
func (c *PanelClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range c.AuthHeaders() {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respData))
	}

	return respData, nil
}

// parseResponse 解析 1Panel 标准响应 { code, message, data }
func parseResponse(data []byte, result interface{}) error {
	var wrapper struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("unmarshal wrapper: %w", err)
	}
	if wrapper.Code != 200 && wrapper.Code != 0 {
		return fmt.Errorf("API error %d: %s", wrapper.Code, wrapper.Message)
	}
	if result == nil || wrapper.Data == nil {
		return nil
	}
	return json.Unmarshal(wrapper.Data, result)
}

// ============================================================
// 数据类型定义（基于 1Panel Agent Swagger 文档）
// ============================================================

// OsInfo 操作系统信息
// GET /dashboard/base/os
type OsInfo struct {
	OS             string `json:"os"`
	Platform       string `json:"platform"`
	PlatformFamily string `json:"platformFamily"`
	KernelVersion  string `json:"kernelVersion"`
	KernelArch     string `json:"kernelArch"`
	DiskSize       int64  `json:"diskSize"`
	PrettyDistro   string `json:"prettyDistro"`
}

// DashboardBase 仪表盘基础信息
// GET /dashboard/base/:ioOption/:netOption
type DashboardBase struct {
	AppInstalledNumber  int               `json:"appInstalledNumber"`
	CPUCores            int               `json:"cpuCores"`
	CPULogicalCores     int               `json:"cpuLogicalCores"`
	CPUMhz              float64           `json:"cpuMhz"`
	CPUModelName        string            `json:"cpuModelName"`
	CronjobNumber       int               `json:"cronjobNumber"`
	CurrentInfo         *DashboardCurrent `json:"currentInfo"`
	DatabaseNumber      int               `json:"databaseNumber"`
	Hostname            string            `json:"hostname"`
	IPv4Addr            string            `json:"ipV4Addr"`
	KernelArch          string            `json:"kernelArch"`
	KernelVersion       string            `json:"kernelVersion"`
	OS                  string            `json:"os"`
	Platform            string            `json:"platform"`
	PlatformFamily      string            `json:"platformFamily"`
	PlatformVersion     string            `json:"platformVersion"`
	PrettyDistro        string            `json:"prettyDistro"`
	SystemProxy         string            `json:"systemProxy"`
	VirtualizationSystem string           `json:"virtualizationSystem"`
	WebsiteNumber       int               `json:"websiteNumber"`
}

// DashboardCurrent 仪表盘实时指标
// GET /dashboard/current/:ioOption/:netOption
type DashboardCurrent struct {
	CPUTotal             int             `json:"cpuTotal"`
	CPUUsed              float64         `json:"cpuUsed"`
	CPUUsedPercent       float64         `json:"cpuUsedPercent"`
	CPUDetailedPercent   []float64       `json:"cpuDetailedPercent"`
	MemoryTotal          int64           `json:"memoryTotal"`
	MemoryUsed           int64           `json:"memoryUsed"`
	MemoryUsedPercent    float64         `json:"memoryUsedPercent"`
	MemoryAvailable      int64           `json:"memoryAvailable"`
	MemoryCache          int64           `json:"memoryCache"`
	MemoryFree           int64           `json:"memoryFree"`
	MemoryShard          int64           `json:"memoryShard"`
	SwapMemoryTotal      int64           `json:"swapMemoryTotal"`
	SwapMemoryUsed       int64           `json:"swapMemoryUsed"`
	SwapMemoryUsedPercent float64        `json:"swapMemoryUsedPercent"`
	SwapMemoryAvailable  int64           `json:"swapMemoryAvailable"`
	IOReadBytes          int64           `json:"ioReadBytes"`
	IOWriteBytes         int64           `json:"ioWriteBytes"`
	IOReadTime           int64           `json:"ioReadTime"`
	IOWriteTime          int64           `json:"ioWriteTime"`
	IOCount              int             `json:"ioCount"`
	LoadUsagePercent     float64         `json:"loadUsagePercent"`
	Load1                float64         `json:"load1"`
	Load5                float64         `json:"load5"`
	Load15               float64         `json:"load15"`
	NetBytesRecv         int64           `json:"netBytesRecv"`
	NetBytesSent         int64           `json:"netBytesSent"`
	Procs                int             `json:"procs"`
	Uptime               int64           `json:"uptime"`
	TimeSinceUptime      string          `json:"timeSinceUptime"`
	ShotTime             string          `json:"shotTime"`
	DiskData             []DiskInfo      `json:"diskData"`
	GPUData              []GPUInfo       `json:"gpuData"`
	TopCPUItems          []Process       `json:"topCPUItems"`
	TopMemItems          []Process       `json:"topMemItems"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Device            string  `json:"device"`
	Path              string  `json:"path"`
	Type              string  `json:"type"`
	Total             int64   `json:"total"`
	Used              int64   `json:"used"`
	Free              int64   `json:"free"`
	UsedPercent       float64 `json:"usedPercent"`
	InodesTotal       int64   `json:"inodesTotal"`
	InodesUsed        int64   `json:"inodesUsed"`
	InodesFree        int64   `json:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent"`
}

// GPUInfo GPU 信息
type GPUInfo struct {
	Index            int    `json:"index"`
	ProductName      string `json:"productName"`
	Temperature      string `json:"temperature"`
	FanSpeed         string `json:"fanSpeed"`
	GPUUtil          string `json:"gpuUtil"`
	MemoryUsage      string `json:"memoryUsage"`
	MemTotal         string `json:"memTotal"`
	MemUsed          string `json:"memUsed"`
	PowerUsage       string `json:"powerUsage"`
	MaxPowerLimit    string `json:"maxPowerLimit"`
	PerformanceState string `json:"performanceState"`
}

// Process 进程信息（Top CPU/Memory）
type Process struct {
	PID     int32   `json:"pid"`
	Name    string  `json:"name"`
	Cmd     string  `json:"cmd"`
	Percent float64 `json:"percent"`
	Memory  int64   `json:"memory"`
	User    string  `json:"user"`
}

// ContainerStatus 容器状态汇总
// GET /containers/status
type ContainerStatus struct {
	ContainerCount      int `json:"containerCount"`
	Running             int `json:"running"`
	Paused              int `json:"paused"`
	Exited              int `json:"exited"`
	Created             int `json:"created"`
	Dead                int `json:"dead"`
	Restarting          int `json:"restarting"`
	Removing            int `json:"removing"`
	ImageCount          int `json:"imageCount"`
	NetworkCount        int `json:"networkCount"`
	VolumeCount         int `json:"volumeCount"`
	ComposeCount        int `json:"composeCount"`
	ComposeTemplateCount int `json:"composeTemplateCount"`
	RepoCount           int `json:"repoCount"`
}

// ContainerInfo 容器详情（与 agent/app/dto/container.go 的 ContainerInfo 保持一致）
// GET /containers/search 返回的数据结构
type ContainerInfo struct {
	ContainerID  string   `json:"containerID"`
	Name         string   `json:"name"`
	ImageID      string   `json:"imageID"`
	ImageName    string   `json:"imageName"`
	CreateTime   string   `json:"createTime"`
	State        string   `json:"state"`
	RunTime      string   `json:"runTime"`    // 运行时长，如 "Up 2 hours"
	Network      []string `json:"network"`    // IP 地址列表
	Ports        []string `json:"ports"`      // 端口映射

	IsFromApp     bool `json:"isFromApp"`
	IsFromCompose bool `json:"isFromCompose"`

	AppName        string   `json:"appName"`
	AppInstallName string   `json:"appInstallName"`
	Websites       []string `json:"websites"`

	IsPinned    bool   `json:"isPinned"`
	Description string `json:"description"`

	// 运行时资源统计（从 /containers/list/stats 合并）
	CPUPercent float64 `json:"cpuPercent"`
	MemUsage   int64   `json:"memUsage"`
	MemLimit   int64   `json:"memLimit"`
}

// ContainerStats 容器资源统计
type ContainerStats struct {
	ContainerID   string  `json:"containerID"`
	CPUTotalUsage int64   `json:"cpuTotalUsage"`
	SystemUsage   int64   `json:"systemUsage"`
	CPUPercent    float64 `json:"cpuPercent"`
	PerCPUUsage   float64 `json:"percpuUsage"`
	MemoryCache   int64   `json:"memoryCache"`
	MemoryUsage   int64   `json:"memoryUsage"`
	MemoryLimit   int64   `json:"memoryLimit"`
	MemoryPercent float64 `json:"memoryPercent"`
}

// MonitorSearchReq 监控历史查询请求（发往 agent 的）
// POST /hosts/monitor/search
type MonitorSearchReq struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Param     string `json:"param"`
	IO        string `json:"io"`
	Network   string `json:"network"`
}

// MonitorSearchResult 监控历史查询结果
// agent 返回的是单个对象 { date: [...], value: [...] }，不是数组
type MonitorSearchResult struct {
	Date  []string          `json:"date"`
	Value []json.RawMessage `json:"value"`
}

// PageContainerReq 容器列表分页查询请求
// POST /containers/search
type PageContainerReq struct {
	Page            int    `json:"page"`
	PageSize        int    `json:"pageSize"`
	Name            string `json:"name"`
	State           string `json:"state"`     // required, oneof: all|created|running|paused|restarting|removing|exited|dead
	OrderBy         string `json:"orderBy"`    // required, oneof: name|createdAt|state
	Order           string `json:"order"`      // required, oneof: null|ascending|descending
	Filters         string `json:"filters"`
	ExcludeAppStore bool   `json:"excludeAppStore"`
}

// PageResult 分页结果
type PageResult struct {
	Items json.RawMessage `json:"items"`
	Total int64           `json:"total"`
}

// ============================================================
// API 方法
// ============================================================

// Ping 测试连接（调用 /dashboard/base/os 验证 API Key 有效性）
func (c *PanelClient) Ping() error {
	_, err := c.doRequest("GET", "/dashboard/base/os", nil)
	return err
}

// GetDashboardOS 获取操作系统信息
// GET /dashboard/base/os
func (c *PanelClient) GetDashboardOS() (*OsInfo, error) {
	data, err := c.doRequest("GET", "/dashboard/base/os", nil)
	if err != nil {
		return nil, err
	}
	var result OsInfo
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDashboardBase 获取仪表盘基础信息（含 currentInfo）
// GET /dashboard/base/:ioOption/:netOption
// ioOption/netOption 为磁盘/网卡过滤选项，传 "-" 表示不限
func (c *PanelClient) GetDashboardBase(ioOption, netOption string) (*DashboardBase, error) {
	path := fmt.Sprintf("/dashboard/base/%s/%s", ioOption, netOption)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var result DashboardBase
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDashboardCurrent 获取当前实时指标
// GET /dashboard/current/:ioOption/:netOption
func (c *PanelClient) GetDashboardCurrent(ioOption, netOption string) (*DashboardCurrent, error) {
	path := fmt.Sprintf("/dashboard/current/%s/%s", ioOption, netOption)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var result DashboardCurrent
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTopCPU 获取 CPU 占用 Top 进程
// GET /dashboard/current/top/cpu
func (c *PanelClient) GetTopCPU() ([]Process, error) {
	data, err := c.doRequest("GET", "/dashboard/current/top/cpu", nil)
	if err != nil {
		return nil, err
	}
	var result []Process
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTopMemory 获取内存占用 Top 进程
// GET /dashboard/current/top/mem
func (c *PanelClient) GetTopMemory() ([]Process, error) {
	data, err := c.doRequest("GET", "/dashboard/current/top/mem", nil)
	if err != nil {
		return nil, err
	}
	var result []Process
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetContainerStatus 获取容器状态汇总
// GET /containers/status
func (c *PanelClient) GetContainerStatus() (*ContainerStatus, error) {
	data, err := c.doRequest("GET", "/containers/status", nil)
	if err != nil {
		return nil, err
	}
	var result ContainerStatus
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListContainers 获取容器列表（分页）
// POST /containers/search
func (c *PanelClient) ListContainers(req PageContainerReq) (*PageResult, error) {
	data, err := c.doRequest("POST", "/containers/search", req)
	if err != nil {
		return nil, err
	}
	var result PageResult
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListContainersSimple 获取容器列表（简化参数，默认查全部，按创建时间降序）
func (c *PanelClient) ListContainersSimple(page, pageSize int) (*PageResult, error) {
	return c.ListContainers(PageContainerReq{
		Page:     page,
		PageSize: pageSize,
		State:    "all",
		OrderBy:  "createdAt",
		Order:    "descending",
	})
}

// ListAllContainers 获取所有容器列表（不分页，最多 1000 个）
func (c *PanelClient) ListAllContainers() (*PageResult, error) {
	return c.ListContainers(PageContainerReq{
		Page:     1,
		PageSize: 1000,
		State:    "all",
		OrderBy:  "createdAt",
		Order:    "descending",
	})
}

// GetContainerStatsList 获取所有运行容器的资源统计
// GET /containers/list/stats
func (c *PanelClient) GetContainerStatsList() ([]ContainerStats, error) {
	data, err := c.doRequest("GET", "/containers/list/stats", nil)
	if err != nil {
		return nil, err
	}
	var result []ContainerStats
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetContainerStats 获取单个容器资源统计
// GET /containers/stats/:id
func (c *PanelClient) GetContainerStats(containerID string) (*ContainerStats, error) {
	data, err := c.doRequest("GET", "/containers/stats/"+containerID, nil)
	if err != nil {
		return nil, err
	}
	var result ContainerStats
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SearchMonitor 查询历史监控数据
// POST /hosts/monitor/search
// param: cpu | memory | load | disk | io | network | gpu
// 返回的是 { date: []string, value: []object } 结构
func (c *PanelClient) SearchMonitor(req MonitorSearchReq) (*MonitorSearchResult, error) {
	data, err := c.doRequest("POST", "/hosts/monitor/search", req)
	if err != nil {
		return nil, err
	}
	var result MonitorSearchResult
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
