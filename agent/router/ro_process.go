package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/masx200/multinodewatchpanel/agent/app/api/v2"
)

type ProcessRouter struct {
}

func (f *ProcessRouter) InitRouter(Router *gin.RouterGroup) {
	processRouter := Router.Group("process")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		processRouter.GET("/ws", baseApi.ProcessWs)
		processRouter.POST("/stop", baseApi.StopProcess)
		processRouter.POST("/listening", baseApi.GetListeningProcess)
		processRouter.POST("/list", baseApi.ListProcesses)
		processRouter.GET("/:pid", baseApi.GetProcessInfoByPID)
	}
}
