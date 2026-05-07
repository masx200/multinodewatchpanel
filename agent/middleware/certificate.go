package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/masx200/multinodewatchpanel/agent/app/api/v2/helper"
	"github.com/masx200/multinodewatchpanel/agent/global"
	"github.com/masx200/multinodewatchpanel/agent/utils/cmd"
	"github.com/masx200/multinodewatchpanel/agent/utils/xpack"
)

func Certificate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.IsMaster {
			c.Next()
			return
		}
		if !xpack.ValidateCertificate(c) {
			CloseDirectly(c)
			return
		}
		conn := c.Request.Header.Get("Connection")
		if conn == "Upgrade" {
			c.Next()
			return
		}
		masterProxyID := c.Request.Header.Get("Proxy-Id")
		proxyID, err := cmd.RunDefaultWithStdoutBashC("cat /etc/1panel/.nodeProxyID")
		if err == nil && len(proxyID) != 0 && strings.TrimSpace(proxyID) != strings.TrimSpace(masterProxyID) {
			helper.InternalServer(c, fmt.Errorf("err proxy id"))
			return
		}
		c.Next()
	}
}

func CloseDirectly(c *gin.Context) {
	hijacker, ok := c.Writer.(http.Hijacker)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	_ = conn.(*net.TCPConn).SetLinger(0)
	conn.Close()
}
