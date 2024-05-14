package main

import (
	"chat/tools"
	"fmt"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	//创建一个访问的路由，并且返回一个Gin服务
	ginServer := tools.GetRouter()

	err := ginServer.Run(":8080")
	if err != nil {
		panic(err)
	}
}
