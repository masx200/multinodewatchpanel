package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/masx200/multinodewatchpanel/agent/app/api/v2/helper"
)

func (b *BaseApi) CheckHealth(c *gin.Context) {
	helper.Success(c)
}
