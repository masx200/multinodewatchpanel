package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/masx200/multinodewatchpanel/agent/app/api/v2"
)

type WebsiteAcmeAccountRouter struct {
}

func (a *WebsiteAcmeAccountRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/acme")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteAcmeAccount)
		groupRouter.POST("", baseApi.CreateWebsiteAcmeAccount)
		groupRouter.POST("/del", baseApi.DeleteWebsiteAcmeAccount)
		groupRouter.POST("/update", baseApi.UpdateWebsiteAcmeAccount)
	}
}
