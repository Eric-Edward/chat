package router

import (
	"chat/services"
	"github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
	//使用 Gin 框架
	ginServer := gin.Default()

	//编写WebSocket 连接的路由
	routerLogin := ginServer.Group("/enter")
	{
		routerLogin.GET("", services.Login)
	}

	// 这个是信息的其他处理，获取更多的历史信息，但是没有前端
	routerMessage := ginServer.Group("/message")
	{
		routerMessage.GET("", services.MessageService)
	}

	return ginServer
}
