package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/masx200/multinodewatchpanel/agent/app/api/v2"
)

type WebsiteCARouter struct {
}

func (a *WebsiteCARouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/ca")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteCA)
		groupRouter.POST("", baseApi.CreateWebsiteCA)
		groupRouter.POST("/del", baseApi.DeleteWebsiteCA)
		groupRouter.POST("/obtain", baseApi.ObtainWebsiteCA)
		groupRouter.POST("/renew", baseApi.RenewWebsiteCA)
		groupRouter.GET("/:id", baseApi.GetWebsiteCA)
		groupRouter.POST("/download", baseApi.DownloadCAFile)
	}
}
