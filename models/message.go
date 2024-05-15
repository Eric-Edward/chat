package models

type Message struct {
	ID       uint   `gorm:"primary_key" json:"id"`
	Time     string `json:"time"`
	Content  string `json:"content"`
	UserName string `json:"user_name"`
}
