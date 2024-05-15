package dao

import (
	"chat/models"
	"chat/tools"
	"math"
	"strconv"
)

// GetMessage 用于获取数据库中的信息
func GetMessage(id string, limit int) ([]models.Message, error) {
	var messages []models.Message
	db := tools.GetDB()
	result := db.Model(&models.Message{}).Where("id <= ?", id).Limit(limit).Order("id desc").Find(&messages)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}
	return messages, nil

}

// GetLatestMessage  程序刚运行时加载到redis中的内容
func GetLatestMessage() ([]models.Message, error) {
	//math.MaxUint>>1-1是为了转换时能够正常表示
	return GetMessage(strconv.Itoa(math.MaxUint>>1-1), 50)
}

// GetLatestMessageByLimit 根据id的编号和需求获取信息
func GetLatestMessageByLimit(limit int) ([]models.Message, error) {
	return GetMessage(strconv.Itoa(math.MaxUint>>1-1), limit)
}
