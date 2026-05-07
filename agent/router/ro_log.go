package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/masx200/multinodewatchpanel/agent/app/api/v2"
)

type LogRouter struct{}

func (s *LogRouter) InitRouter(Router *gin.RouterGroup) {
	operationRouter := Router.Group("logs")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		operationRouter.GET("/system/files", baseApi.GetSystemFiles)
		operationRouter.POST("/tasks/search", baseApi.PageTasks)
		operationRouter.GET("/tasks/executing/count", baseApi.CountExecutingTasks)
	}
}
