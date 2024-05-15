package global

import (
	"chat/dao"
	"chat/models"
	"chat/tools"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"strconv"
	"sync"
	"time"
)

type ILocke struct {
	mu sync.Mutex
	id uint
}

type Connection struct {
	Username string
	WS       *websocket.Conn
	FromWS   chan models.Message
	ToWS     chan models.Message
	Close    chan struct{}
}

// AllConnections 这个用于存储所有的连接到服务器的websocket,且每一个连接都是一个Connection对象
var AllConnections sync.Map

// 使用redis时，给信息一个id
var latest *ILocke

var timer *time.Ticker

func init() {
	AllConnections = sync.Map{}
	latest = &ILocke{
		mu: sync.Mutex{},
		id: 0,
	}
	timer = time.NewTicker(time.Minute * 2)
	preLoadMessage()

	go persistData()
}

// persistData 这个函数用于将redis中的数据持久化到mysql中保存
func persistData() {
	for {
		select {
		case <-timer.C:

		default:
			//不做什么，防止阻塞
		}
	}
}

// preLoadMessage 在服务运行前，提前将一些数据放到redis中。
func preLoadMessage() {
	rdb := tools.GetRedis()
	messages, err := dao.GetLatestMessage()
	if err != nil {
		fmt.Println(err)
	}
	latest.mu.Lock()
	latest.id = uint(len(messages))
	latest.mu.Unlock()
	for i := 0; i < len(messages); i++ {
		marshal, e := json.Marshal(messages[i])
		if e != nil {
			marshal, e = json.Marshal(messages[i])
			if e != nil {
				break
			}
		}
		rdb.Set(context.Background(), strconv.Itoa(int(messages[i].ID)), string(marshal), time.Hour)
	}
}

// ReceiveMessage 这个函数是用来为客户运行一个从WS接受消息的goroutine
func (conn *Connection) ReceiveMessage() {
	for {
		select {
		//从自己的同步接受携程的管道中获取信息
		case msg := <-conn.FromWS:
			rdb := tools.GetRedis()

			//给每个信息一个独立的id
			latest.mu.Lock()
			msg.ID = latest.id
			latest.id++
			latest.mu.Unlock()
			marshal, err := json.Marshal(msg)
			if err != nil {
				break
			}

			//使用string类型来进行保存，相比来说，空间浪费应该会小一点
			rdb.Set(context.Background(), string(strconv.Itoa(int(msg.ID))), marshal, time.Hour)

			//循环，给聊天室中的每个人都返回信息
			AllConnections.Range(func(key, value interface{}) bool {
				if key.(string) != conn.Username {
					value.(*Connection).ToWS <- msg
				}
				return true
			})
		case <-conn.Close:
			//这个是用于ws断开连接时，关闭该ws对应的goroutines
			return
		default:
		}
	}
}

// SendMessage 这个函数用来为客户运行一个发送至WS的goroutine
func (conn *Connection) SendMessage() {
	timer := time.NewTicker(time.Second * 10)
	for {
		select {
		case msg := <-conn.ToWS:
			err := conn.WS.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
			}
		case <-timer.C:
			_ = conn.WS.WriteMessage(websocket.PingMessage, []byte{})

		case <-conn.Close:
			timer.Stop()
			return
		default:

		}
	}
}
