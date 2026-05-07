package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/masx200/multinodewatchpanel/core/app/api/v2"
)

type BaseRouter struct{}

func (s *BaseRouter) InitRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("auth")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		baseRouter.GET("/captcha", baseApi.Captcha)
		baseRouter.POST("/passkey/begin", baseApi.PasskeyBeginLogin)
		baseRouter.POST("/passkey/finish", baseApi.PasskeyFinishLogin)
		baseRouter.POST("/mfalogin", baseApi.MFALogin)
		baseRouter.POST("/login", baseApi.Login)
		baseRouter.POST("/logout", baseApi.LogOut)
		baseRouter.GET("/setting", baseApi.GetLoginSetting)
		baseRouter.GET("/welcome", baseApi.GetWelcomePage)
	}
}
