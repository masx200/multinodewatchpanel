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

// PanelClient 1Panel API 客户端
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
	baseURL := fmt.Sprintf("%s://%s:%d", scheme, host, port)

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

// parseResponse 解析 1Panel 标准响应
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

// ---- 数据类型定义 ----

type OsInfo struct {
	OS             string `json:"os"`
	Platform       string `json:"platform"`
	PlatformFamily string `json:"platformFamily"`
	KernelVersion  string `json:"kernelVersion"`
	KernelArch     string `json:"kernelArch"`
	Hostname       string `json:"hostname"`
	Uptime         uint64 `json:"uptime"`
}

type DashboardBase struct {
	CPULogicalCount  int    `json:"cpuLogicalCount"`
	CPUPhysicalCount int    `json:"cpuPhysicalCount"`
	MemoryTotal      uint64 `json:"memoryTotal"`
	SwapTotal        uint64 `json:"swapTotal"`
	DiskTotal        uint64 `json:"diskTotal"`
	LoadAverage      string `json:"loadAverage"`
	SystemVersion    string `json:"systemVersion"`
	DockerVersion    string `json:"dockerVersion"`
}

type DashboardCurrent struct {
	CPUPercent      float64 `json:"cpuPercent"`
	MemoryUsed      uint64  `json:"memoryUsed"`
	MemoryPercent   float64 `json:"memoryPercent"`
	SwapUsed        uint64  `json:"swapUsed"`
	DiskUsed        uint64  `json:"diskUsed"`
	DiskPercent     float64 `json:"diskPercent"`
	NetworkUpload   float64 `json:"networkUpload"`
	NetworkDownload float64 `json:"networkDownload"`
	IORead          float64 `json:"ioRead"`
	IOWrite         float64 `json:"ioWrite"`
	Load1           float64 `json:"load1"`
	Load5           float64 `json:"load5"`
	Load15          float64 `json:"load15"`
}

type Process struct {
	PID    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory float32 `json:"memory"`
}

type ContainerStatus struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Paused  int `json:"paused"`
}

type ContainerInfo struct {
	ContainerID string  `json:"containerID"`
	Name        string  `json:"name"`
	ImageName   string  `json:"imageName"`
	State       string  `json:"state"`
	Status      string  `json:"status"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemUsage    float64 `json:"memUsage"`
	MemLimit    float64 `json:"memLimit"`
}

type ContainerStats struct {
	CPUPercent float64 `json:"cpuPercent"`
	MemUsage   float64 `json:"memUsage"`
	MemLimit   float64 `json:"memLimit"`
	MemPercent float64 `json:"memPercent"`
	IORead     float64 `json:"ioRead"`
	IOWrite    float64 `json:"ioWrite"`
	NetInput   float64 `json:"netInput"`
	NetOutput  float64 `json:"netOutput"`
}

type MonitorData struct {
	Date  interface{} `json:"date"`
	Value interface{} `json:"value"`
}

type ListContainerReq struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Name     string `json:"name"`
}

type MonitorSearchReq struct {
	Param     string `json:"param"`
	Info      string `json:"info"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// ---- API 方法 ----

// Ping 测试连接
func (c *PanelClient) Ping() error {
	_, err := c.doRequest("POST", "/api/v2/dashboard/os", nil)
	return err
}

// GetDashboardOS 获取操作系统信息
func (c *PanelClient) GetDashboardOS() (*OsInfo, error) {
	data, err := c.doRequest("POST", "/api/v2/dashboard/os", nil)
	if err != nil {
		return nil, err
	}
	var result OsInfo
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDashboardBase 获取基础信息
func (c *PanelClient) GetDashboardBase(ioOption, netOption string) (*DashboardBase, error) {
	payload := map[string]string{"ioOption": ioOption, "netOption": netOption}
	data, err := c.doRequest("POST", "/api/v2/dashboard/base/search", payload)
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
func (c *PanelClient) GetDashboardCurrent(ioOption, netOption string) (*DashboardCurrent, error) {
	payload := map[string]string{"ioOption": ioOption, "netOption": netOption}
	data, err := c.doRequest("POST", "/api/v2/dashboard/current", payload)
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
func (c *PanelClient) GetTopCPU() ([]Process, error) {
	data, err := c.doRequest("GET", "/api/v2/dashboard/process/top/cpu", nil)
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
func (c *PanelClient) GetTopMemory() ([]Process, error) {
	data, err := c.doRequest("GET", "/api/v2/dashboard/process/top/mem", nil)
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
func (c *PanelClient) GetContainerStatus() (*ContainerStatus, error) {
	data, err := c.doRequest("GET", "/api/v2/containers/status", nil)
	if err != nil {
		return nil, err
	}
	var result ContainerStatus
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListContainers 获取容器列表
func (c *PanelClient) ListContainers(req ListContainerReq) ([]ContainerInfo, error) {
	data, err := c.doRequest("POST", "/api/v2/containers/search", req)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Items []ContainerInfo `json:"items"`
		Total int64           `json:"total"`
	}
	if err := parseResponse(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Items, nil
}

// GetContainerStats 获取单个容器资源统计
func (c *PanelClient) GetContainerStats(containerID string) (*ContainerStats, error) {
	data, err := c.doRequest("GET", "/api/v2/containers/stats/"+containerID, nil)
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
func (c *PanelClient) SearchMonitor(req MonitorSearchReq) ([]MonitorData, error) {
	data, err := c.doRequest("POST", "/api/v2/monitor/search", req)
	if err != nil {
		return nil, err
	}
	var result []MonitorData
	if err := parseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListContainersSimple 获取容器列表（简化参数）
func (c *PanelClient) ListContainersSimple(page, pageSize int) ([]ContainerInfo, error) {
	return c.ListContainers(ListContainerReq{Page: page, PageSize: pageSize})
}
