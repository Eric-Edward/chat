package tools

import (
	"chat/services"
	"github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
	//使用 Gin 框架
	ginServer := gin.Default()

	//编写WebSocket 连接的路由
	routerLogin := ginServer.Group("ws://")
	{
		routerLogin.GET("", services.Login)
	}

	return ginServer
}
