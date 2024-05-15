package main

import (
	_ "chat/global"
	"chat/router"
	"fmt"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	//创建一个访问的路由，并且返回一个Gin服务
	ginServer := router.GetRouter()

	//运行在8080端口
	err := ginServer.Run("127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
}
