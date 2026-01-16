package jjf4

import (
	"context"
	"fmt"

	redis2 "github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var redisDb *redis2.Client
var mysqlDb *gorm.DB

func InitRedis(ctx context.Context) *redis2.Client {
	if redisDb == nil {
		rdb := redis2.NewClient(&redis2.Options{
			Addr:          fmt.Sprintf("%s:%d", "127.0.0.1", 6379),
			Protocol:      2,
			UnstableResp3: true,
		})

		if _, err := rdb.Ping(ctx).Result(); err != nil {
			panic(fmt.Sprintf("Redis连接失败: %v", err))
		}
		return rdb
	} else {
		return redisDb
	}
}

func InitMysql(ctx context.Context) *gorm.DB {
	if mysqlDb == nil {
		dsn := "root:root@tcp(127.0.0.1:3306)/lw_match?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			PrepareStmt: true,
		})
		if err != nil {
			panic(err)
		}
		return db
	}
	return mysqlDb
}
