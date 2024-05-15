package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func Test(t *testing.T) {
	db, err := gorm.Open(mysql.Open(`root:Tsinghua@tcp(127.0.0.1:3306)/test?charset=utf8`), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&Message{})

}
