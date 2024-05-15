package global

import (
	"chat/dao"
	"chat/models"
	"chat/tools"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

// ILock 这个结构体用户维护当前信息的最新的id
type ILock struct {
	mu sync.Mutex
	id uint
}

// Connection 用户表示一个ws连接
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
var latest *ILock

// 这个timer是一个用于redis数据持久话的定时器
var timer *time.Ticker

func init() {
	AllConnections = sync.Map{}
	latest = &ILock{
		mu: sync.Mutex{},
		id: 0,
	}
	timer = time.NewTicker(time.Minute * 2)
	preLoadMessage()

	go persistData()
}

// persistData 这个函数用于将redis中的数据持久化到mysql中保存
func persistData() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	for {
		select {
		//每间隔2分钟就讲redis中最新的数据保存至mysql中，并且将上一个2分钟的数据删除
		case <-timer.C:

			// 获取数据库中目前保存的最新的信息的id
			var maxId uint
			db := tools.GetDB()
			tx := db.Model(&models.Message{}).Select("id").Order("id desc").Limit(1).First(&maxId)
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				maxId = 0
			}

			//或者redis中所有的信息
			rdb := tools.GetRedis()
			keys := rdb.Keys(context.Background(), "*").Val()
			var needPersist []models.Message
			for i := range keys {
				key, err := strconv.Atoi(keys[i])

				// 持久话最新的，删除上一个2分钟的信息，但是同时保证redis中仍然有信息
				if err == nil && uint(key) <= maxId && len(keys) >= 10 {
					rdb.Unlink(context.Background(), keys[i])
				} else if uint(key) > maxId {
					var msg models.Message
					err = json.Unmarshal([]byte(rdb.Get(context.Background(), keys[i]).Val()), &msg)
					if err != nil {
						panic(err)
					}
					needPersist = append(needPersist, msg)
				}
			}

			//使用事务提交需要更新的部分
			if len(needPersist) > 0 {
				begin := db.Begin()
				result := begin.Table("messages").Create(needPersist)
				if result.Error != nil {
					begin.Rollback()
					panic(result.Error)
				}
				begin.Commit()
			}
		default:
			//添加一个默认。
		}
	}
}

// preLoadMessage 在服务运行前，提前将一些数据放到redis中。
func preLoadMessage() {
	rdb := tools.GetRedis()
	rdb.FlushDB(context.Background())
	messages, err := dao.GetLatestMessage()
	if err != nil {
		fmt.Println(err)
	}

	// 互斥的拿到当前最新信息的id
	latest.mu.Lock()
	latest.id = uint(len(messages))
	latest.mu.Unlock()

	// 提前将聊天室中的信息读到redis中，方便查询
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
	heart := time.NewTicker(time.Second * 10)
	for {
		select {
		// 转发给用户的ws
		case msg := <-conn.ToWS:
			err := conn.WS.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
			}
			//心跳
		case <-heart.C:
			_ = conn.WS.WriteMessage(websocket.PingMessage, []byte{})

			//用户关闭计时器和退出当前携程
		case <-conn.Close:
			heart.Stop()
			return
		default:

		}
	}
}

// GetLatestId 这个函数用于返回当前服务器中维护的id最大值
func (conn *Connection) GetLatestId() uint {
	return latest.id
}
