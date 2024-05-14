package services

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 使用gorilla包将当前的http请求升级为WebSocket连接
var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Login(c *gin.Context) {
	//请求升级为WebSocket
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(503, gin.H{
			"message": "Failed to upgrade to websocket",
		})
		return
	}
	defer func() {
		_ = ws.Close()
	}()

	c.JSON(200, gin.H{
		"message": "success",
	})
}
