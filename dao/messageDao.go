package dao

import (
	"chat/models"
	"chat/tools"
	"math"
	"strconv"
)

func GetMessage(id string) ([]models.Message, error) {
	var messages []models.Message
	db := tools.GetDB()
	result := db.Model(&models.Message{}).Where("id <= ?", id).Limit(50).Order("id desc").Find(&messages)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}
	return messages, nil

}

func GetLatestMessage() ([]models.Message, error) {
	return GetMessage(strconv.Itoa(math.MaxUint>>1 - 1))
}
