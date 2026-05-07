package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/masx200/multinodewatchpanel/core/app/api/v2"
	"github.com/masx200/multinodewatchpanel/core/middleware"
)

type GroupRouter struct {
}

func (a *GroupRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("groups").
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("", baseApi.CreateGroup)
		groupRouter.POST("/del", baseApi.DeleteGroup)
		groupRouter.POST("/update", baseApi.UpdateGroup)
		groupRouter.POST("/search", baseApi.ListGroup)
	}
}
