package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/masx200/multinodewatchpanel/core/app/api/v2/helper"
	"github.com/masx200/multinodewatchpanel/core/app/repo"
	"github.com/masx200/multinodewatchpanel/core/buserr"
	"github.com/masx200/multinodewatchpanel/core/constant"
	"github.com/masx200/multinodewatchpanel/core/global"
)

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiReq := c.GetBool("API_AUTH")
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") || apiReq {
			c.Next()
			return
		}

		psession, err := global.SESSION.Get(c)
		if err != nil {
			errItem := err.Error()
			if errItem == "ErrSessionDataFormat" || errItem == "ErrSessionDataNotFound" {
				helper.BadAuth(c, "ErrNotLogin", buserr.New(errItem))
				return
			}
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		settingRepo := repo.NewISettingRepo()
		sessionTimeout, err := settingRepo.GetValueByKey("SessionTimeout")
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
			return
		}
		lifeTime, _ := strconv.Atoi(sessionTimeout)
		ssl, err := settingRepo.GetValueByKey("SSL")
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
			return
		}
		if _, err := global.SESSION.RefreshIfNeeded(c, psession, ssl == constant.StatusEnable, lifeTime); err != nil {
			errItem := err.Error()
			if errItem == "ErrSessionDataFormat" || errItem == "ErrSessionDataNotFound" {
				helper.BadAuth(c, "ErrNotLogin", buserr.New(errItem))
				return
			}
			global.LOG.Warnf("refresh session failed, path=%s, err=%v", c.Request.URL.Path, err)
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		c.Next()
	}
}
