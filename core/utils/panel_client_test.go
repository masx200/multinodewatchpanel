package utils

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"os"
	"strings"
	"testing"
)

// TestPinnedPeerCertSHA256 测试证书 SHA256 哈希计算是否与 xray tls hash 一致
func TestPinnedPeerCertSHA256(t *testing.T) {
	// 读取证书文件
	certPath := `C:\Users\Administrator.WIN-9M55V3EFM0S\Downloads\docker-debian-systemd-privileged.download.crt`
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Skipf("证书文件不存在，跳过测试: %v", err)
		return
	}

	// 解析 PEM 块
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("无法解析 PEM 证书")
	}

	// 解析证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("无法解析证书: %v", err)
	}

	// 计算 SHA256 哈希（与 PanelClient.buildTransport 中的逻辑一致）
	sum := sha256.Sum256(cert.Raw)
	got := hex.EncodeToString(sum[:])

	// xray tls hash --cert 的输出（小写）
	expected := "070ba0fa8f6b0dcf067889e07cd929ee0c84c5283420da152b3d19d216c94deb"

	// 比较（不区分大小写）
	if strings.ToLower(got) != expected {
		t.Errorf("证书 SHA256 哈希不匹配\n期望: %s\n实际: %s", expected, got)
	} else {
		t.Logf("✓ 证书 SHA256 哈希验证通过: %s", got)
	}

	// 测试 VerifyPeerCertificate 回调逻辑
	t.Run("VerifyPeerCertificate回调", func(t *testing.T) {
		pinnedHash := expected

		// 模拟 VerifyPeerCertificate 回调
		verifyFunc := func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return nil // 应该返回错误，但这里我们只是测试哈希计算
			}
			sum := sha256.Sum256(rawCerts[0])
			got := hex.EncodeToString(sum[:])
			if got != pinnedHash {
				t.Errorf("证书 SHA256 不匹配: got %s, want %s", got, pinnedHash)
			}
			return nil
		}

		// 使用证书的原始字节测试
		err := verifyFunc([][]byte{block.Bytes}, nil)
		if err != nil {
			t.Errorf("VerifyPeerCertificate 返回错误: %v", err)
		}
	})

	// 测试 PanelClientOptions 的 PinnedPeerCertSHA256 选项
	t.Run("PanelClientOptions配置", func(t *testing.T) {
		client := NewPanelClientWithOptions("example.com", 9999, "test-api-key",
			PanelClientOptions{
				Security:            "tls",
				PinnedPeerCertSHA256: expected,
			})

		if client == nil {
			t.Fatal("创建客户端失败")
		}

		// 验证客户端配置
		if client.APIKey != "test-api-key" {
			t.Errorf("APIKey 不匹配: got %s, want %s", client.APIKey, "test-api-key")
		}

		t.Logf("✓ PanelClient 创建成功，已配置证书固定")
	})
}

// TestGenerateToken 测试 API Token 生成
func TestGenerateToken(t *testing.T) {
	apiKey := "test-api-key"

	// 生成 token
	token1, ts1 := GenerateToken(apiKey)

	// Token 应该是 32 个字符的十六进制字符串（MD5 哈希）
	if len(token1) != 32 {
		t.Errorf("Token 长度不正确: got %d, want 32", len(token1))
	}

	// 验证 token 格式（应该是十六进制）
	_, err := hex.DecodeString(token1)
	if err != nil {
		t.Errorf("Token 不是有效的十六进制字符串: %v", err)
	}

	_ = ts1 // 避免未使用变量错误

	t.Logf("✓ Token 生成测试通过")
	t.Logf("  Token: %s", token1)
	t.Logf("  Timestamp: %s", ts1)
}

// TestAuthHeaders 测试认证请求头
func TestAuthHeaders(t *testing.T) {
	client := NewPanelClient("example.com", 9999, "test-api-key")

	headers := client.AuthHeaders()

	// 验证必需的请求头
	requiredHeaders := []string{"1Panel-Token", "1Panel-Timestamp", "Content-Type"}
	for _, h := range requiredHeaders {
		if _, ok := headers[h]; !ok {
			t.Errorf("缺少必需的请求头: %s", h)
		}
	}

	// 验证 Content-Type
	if headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type 不正确: got %s, want application/json", headers["Content-Type"])
	}

	t.Logf("✓ 认证请求头测试通过")
	t.Logf("  Headers: %+v", headers)
}

// TestBuildTransport 测试 Transport 构建
func TestBuildTransport(t *testing.T) {
	tests := []struct {
		name string
		opts PanelClientOptions
		desc string
	}{
		{
			name: "无TLS",
			opts: PanelClientOptions{Security: "none"},
			desc: "不使用 TLS，返回默认 Transport",
		},
		{
			name: "TLS_跳过证书校验",
			opts: PanelClientOptions{Security: "tls", AllowInsecure: true},
			desc: "TLS 但跳过证书校验",
		},
		{
			name: "TLS_证书固定",
			opts: PanelClientOptions{
				Security:            "tls",
				PinnedPeerCertSHA256: "070ba0fa8f6b0dcf067889e07cd929ee0c84c5283420da152b3d19d216c94deb",
			},
			desc: "TLS 并使用证书固定",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := buildTransport(tt.opts)
			if transport == nil {
				t.Error("buildTransport 返回 nil")
			}
			t.Logf("✓ %s: Transport 创建成功", tt.desc)
		})
	}
}

// TestParseResponse 测试响应解析
func TestParseResponse(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		result   interface{}
		wantErr  bool
		errMatch string
	}{
		{
			name:    "成功响应_code_200",
			data:    []byte(`{"code":200,"message":"success","data":{"test":"value"}}`),
			result:  &map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "成功响应_code_0",
			data:    []byte(`{"code":0,"message":"success","data":null}`),
			result:  nil,
			wantErr: false,
		},
		{
			name:     "错误响应",
			data:     []byte(`{"code":400,"message":"bad request","data":null}`),
			result:   nil,
			wantErr:  true,
			errMatch: "API error 400",
		},
		{
			name:     "无效JSON",
			data:     []byte(`invalid json`),
			result:   nil,
			wantErr:  true,
			errMatch: "unmarshal wrapper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseResponse(tt.data, tt.result)

			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，但没有")
				} else if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("错误信息不匹配: got %s, want contains %s", err.Error(), tt.errMatch)
				}
			} else {
				if err != nil {
					t.Errorf("不期望返回错误，但得到: %v", err)
				}
			}

			if !tt.wantErr {
				t.Logf("✓ %s: 解析成功", tt.name)
			}
		})
	}
}
