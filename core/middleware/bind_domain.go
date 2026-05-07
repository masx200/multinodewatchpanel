package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/masx200/multinodewatchpanel/core/app/api/v2/helper"
	"github.com/masx200/multinodewatchpanel/core/app/repo"
)

func BindDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		localRequest := c.GetBool("LOCAL_REQUEST")
		if localRequest {
			c.Next()
			return
		}
		settingRepo := repo.NewISettingRepo()
		bindDomain, err := settingRepo.GetValueByKey("BindDomain")
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		if len(bindDomain) == 0 {
			c.Next()
			return
		}
		domains := c.Request.Host
		parts := strings.Split(c.Request.Host, ":")
		if len(parts) > 0 {
			domains = parts[0]
		}

		if domains != bindDomain {
			code := LoadErrCode()
			helper.ErrWithHtml(c, code, "err_domain")
			return
		}
		c.Next()
	}
}
