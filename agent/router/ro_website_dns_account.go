package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/masx200/multinodewatchpanel/agent/app/api/v2"
)

type WebsiteDnsAccountRouter struct {
}

func (a *WebsiteDnsAccountRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/dns")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteDnsAccount)
		groupRouter.POST("", baseApi.CreateWebsiteDnsAccount)
		groupRouter.POST("/update", baseApi.UpdateWebsiteDnsAccount)
		groupRouter.POST("/del", baseApi.DeleteWebsiteDnsAccount)
	}
}
