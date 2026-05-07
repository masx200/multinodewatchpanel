package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/masx200/multinodewatchpanel/core/app/api/v2"
	"github.com/masx200/multinodewatchpanel/core/middleware"
)

type ScriptRouter struct{}

func (s *ScriptRouter) InitRouter(Router *gin.RouterGroup) {
	scriptRouter := Router.Group("script").
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v2.ApiGroupApp.BaseApi
	{
		scriptRouter.POST("", baseApi.CreateScript)
		scriptRouter.POST("/search", baseApi.SearchScript)
		scriptRouter.POST("/del", baseApi.DeleteScript)
		scriptRouter.POST("/update", baseApi.UpdateScript)
		scriptRouter.POST("/sync", baseApi.SyncScript)
		scriptRouter.GET("/run", baseApi.RunScript)
	}
}
