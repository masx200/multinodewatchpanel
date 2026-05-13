package model

import "time"

type HostNode struct {
	BaseModel
	Name      string     `gorm:"not null;size:64" json:"name"`
	Host      string     `gorm:"not null;size:255" json:"host"`
	Port      int        `gorm:"not null;default:9999" json:"port"`
	APIKey    string     `gorm:"column:api_key;not null;size:128" json:"apiKey"`
	Status    int        `gorm:"default:0" json:"status"` // 0=offline 1=online 2=failed
	LastSeen  *time.Time `json:"lastSeen"`
	Tags      string     `gorm:"size:255" json:"tags"`
	IsDefault bool       `gorm:"default:false" json:"isDefault"`
	// 服务名称（路径前缀，如 SNI/ServerName）
	ServerName string `gorm:"column:server_name;size:255;default:''" json:"serverName"`
	// 传输层安全："none" 或 "tls"
	Security string `gorm:"column:security;size:16;default:'none'" json:"security"`
	// 允许不安全连接（跳过证书校验）
	AllowInsecure bool `gorm:"column:allow_insecure;default:false" json:"allowInsecure"`
	// 远程服务器证书 SHA256 散列（hex，大小写不敏感）
	PinnedPeerCertSHA256 string `gorm:"column:pinned_peer_cert_sha256;size:64;default:''" json:"pinnedPeerCertSha256"`
}

func (h HostNode) TableName() string {
	return "host_nodes"
}
