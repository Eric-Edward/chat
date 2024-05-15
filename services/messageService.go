package services

import (
	"chat/dao"
	"chat/models"
	"chat/tools"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strconv"
)

// MessageService 这个函数用于模拟前端请求更多数据
func MessageService(c *gin.Context) {
	//获取需要开始请求的信息的id
	sid := c.Query("id")
	id, _ := strconv.Atoi(sid)
	message, err := GetMessageList(uint(id), 50)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "获取历史信息失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": message,
	})
	return
}

// GetMessageList 这个函数使用redis和mysql的方式进行查询，如果redis存在则查redis，不存在查mysql
func GetMessageList(sid uint, size uint) (messages []models.Message, err error) {
	rdb := tools.GetRedis()
	for i := sid; i >= 0 && i > sid-size; i-- {
		var msg models.Message
		val := rdb.Get(context.Background(), strconv.Itoa(int(i)))

		// redis中存在
		if !errors.Is(val.Err(), redis.Nil) {
			err = json.Unmarshal([]byte(val.Val()), &msg)
			if err != nil {
				return nil, err
			}
			messages = append(messages, msg)
		} else {
			//redis不存在，去访问mysql，并且直接返回
			message, e := dao.GetMessage(strconv.Itoa(int(i)), int(size-(sid-i+1)))
			if e == nil {
				messages = append(messages, message...)
			}
			return
		}

	}
	return
}
