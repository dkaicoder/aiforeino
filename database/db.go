package database

import (
	"context"
	"fmt"
	"log"

	redis2 "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var RedisDb *redis2.Client
var MysqlDb *gorm.DB
var kafkaConn *kafka.Conn

func InitRedis(ctx context.Context) *redis2.Client {
	if RedisDb == nil {
		rdb := redis2.NewClient(&redis2.Options{
			Addr:          fmt.Sprintf("%s:%d", "127.0.0.1", 6379),
			Protocol:      2,
			UnstableResp3: true,
		})

		if _, err := rdb.Ping(ctx).Result(); err != nil {
			panic(fmt.Sprintf("Redis连接失败: %v", err))
		}
		RedisDb = rdb
		return RedisDb
	} else {
		return RedisDb
	}
}

func InitMysql(ctx context.Context) *gorm.DB {
	if MysqlDb == nil {
		dsn := "root:root@tcp(127.0.0.1:3306)/lw_match?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			PrepareStmt: true,
		})
		if err != nil {
			panic(err)
		}
		MysqlDb = db
		return db
	}
	return MysqlDb
}

func InitKafkaForProducer(ctx context.Context) *kafka.Conn {
	if kafkaConn == nil {
		topic := "my-topic"
		partition := 0
		conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
		if err != nil {
			log.Fatal("failed to dial leader:", err)
		}
		return conn
	}
	return kafkaConn
}

func InitKafkaForConsumer(ctx context.Context) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "my-topic",
		Partition: 0,
		MaxBytes:  10e6, // 10MB
		GroupID:   "my-first-kafka-consumer-group",
	})
	return r
}
