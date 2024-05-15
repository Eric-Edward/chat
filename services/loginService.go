package services

import (
	"chat/global"
	"chat/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

// 使用gorilla包将当前的http请求升级为WebSocket连接
var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Login(c *gin.Context) {
	username := c.Query("username")

	//请求升级为WebSocket
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	connection := global.Connection{
		Username: username,
		WS:       ws,
		FromWS:   make(chan models.Message),
		ToWS:     make(chan models.Message),
		Close:    make(chan struct{}),
	}

	//如果这之前登录过，那么先把之前登录的踢掉
	preConn, ok := global.AllConnections.Load(username)
	if ok {
		Logout(preConn.(*global.Connection))
	}
	global.AllConnections.Store(username, &connection)

	//用一个单独的goroutine来接受用户发过来的信息
	go func(conn *global.Connection) {
		defer func() {
			Logout(conn)
		}()
		for {
			mType, p, e := ws.ReadMessage()
			if mType == websocket.CloseMessage || e != nil {
				Logout(conn)
				return
			}
			if mType == websocket.TextMessage || mType == websocket.BinaryMessage {
				conn.FromWS <- models.Message{
					ID:       0,
					UserName: conn.Username,
					Time:     time.Now().Format("2006-01-02 15:04:05"),
					Content:  string(p),
				}
			}
		}
	}(&connection)
	go connection.ReceiveMessage()
	go connection.SendMessage()

	//登录成功逻辑
	AfterLogin(&connection)

	c.JSON(200, gin.H{
		"message": "login successfully",
	})
}

// Logout 这个函数处理用户退出时的逻辑
func Logout(conn *global.Connection) {
	//退出时先从连接的集合中删除
	global.AllConnections.Delete(conn.Username)

	//这两个信号量是用来控制每个用户运行的接受和发送携程的
	conn.Close <- struct{}{}
	conn.Close <- struct{}{}

	//关闭当前连接打开的管道，以免内存泄漏
	close(conn.FromWS)
	close(conn.ToWS)

	//最后关闭ws
	_ = conn.WS.Close()
}

// AfterLogin 登录成功之后发送登录成功，并且发送聊天室内最新的10条信息。
// 如果redis不够10条的话，就去看mysql。然后剩下的也不看redis了。
func AfterLogin(conn *global.Connection) {
	conn.ToWS <- models.Message{
		ID:       0,
		UserName: "system",
		Time:     time.Now().Format("2006-01-02 15:04:05"),
		Content:  "登录成功",
	}
	id := conn.GetLatestId()
	message, err := GetMessageList(id, 10)
	if err != nil {
		fmt.Println(err)
	}
	// 反序返回给用户，但是这里可能会有一个问题。可能新的消息比查询的消息先到用户手里。这就有可能导致
	// 用户看到的信息是乱序的。所以需要在前端根据序号排序后再显示
	for i := len(message) - 1; i >= 0; i-- {
		conn.ToWS <- message[i]
	}
}
