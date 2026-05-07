package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/gin-gonic/gin"
)

type NodeRouter struct{}

func (n *NodeRouter) InitRouter(Router *gin.RouterGroup) {
	nodeRouter := Router.Group("nodes").
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())

	baseApi := v2.ApiGroupApp.BaseApi
	{
		nodeRouter.GET("", baseApi.ListNodes)
		nodeRouter.POST("", baseApi.CreateNode)
		nodeRouter.PUT("/:id", baseApi.UpdateNode)
		nodeRouter.DELETE("/:id", baseApi.DeleteNode)
		nodeRouter.POST("/test", baseApi.TestNodeConnection)
		nodeRouter.GET("/:id/dashboard", baseApi.GetNodeDashboard)
		nodeRouter.GET("/:id/containers", baseApi.GetNodeContainers)
		nodeRouter.POST("/monitor/search", baseApi.SearchMonitorHistory)
	}
}
