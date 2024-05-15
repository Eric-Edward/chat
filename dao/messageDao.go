package dao

import (
	"chat/models"
	"chat/tools"
	"math"
	"strconv"
)

func GetMessage(id string, limit int) ([]models.Message, error) {
	var messages []models.Message
	db := tools.GetDB()
	result := db.Model(&models.Message{}).Where("id <= ?", id).Limit(limit).Order("id desc").Find(&messages)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}
	return messages, nil

}

func GetLatestMessage() ([]models.Message, error) {
	return GetMessage(strconv.Itoa(math.MaxUint>>1-1), 50)
}

func GetLatestMessageByLimit(limit int) ([]models.Message, error) {
	return GetMessage(strconv.Itoa(math.MaxUint>>1-1), limit)
}
