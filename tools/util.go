package tools

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var redisDB *redis.Client
var mysqlDB *gorm.DB

func init() {
	db, err := gorm.Open(mysql.Open(`root:Tsinghua@tcp(127.0.0.1:3306)/test?charset=utf8`), &gorm.Config{})
	//当err不为nil的时候，持续获取连接
	if err != nil {
		db, err = gorm.Open(mysql.Open(`root:Tsinghua@tcp(127.0.0.1:3306)/test?charset=utf8`), &gorm.Config{})
		if err != nil {
			panic("open mysql fail")
		}
		mysqlDB = db
	}
	mysqlDB = db

	redisDB = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "Tsinghua",
	})

}

func GetDB() *gorm.DB {
	return mysqlDB
}

func GetRedis() *redis.Client {
	return redisDB
}
