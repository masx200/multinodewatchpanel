//go:build !xpack && !xpackee

package xpack

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/masx200/multinodewatchpanel/core/utils/ssh"
)

func Proxy(c *gin.Context, currentNode string) {}

func ProxyDocker(proxyURL string) error { return nil }

func UpdateGroup(name string, group, newGroup uint) error { return nil }

func CheckBackupUsed(name string) error { return nil }

func LoadRequestTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       15 * time.Second,
	}
}

func LoadNodeInfo(currentNode string) (*ssh.ConnInfo, string, error) {
	return nil, "", nil
}

func Sync(dataType string) error { return nil }

func AutoUpgradeWithMaster() {}
