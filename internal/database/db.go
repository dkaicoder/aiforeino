package database

import (
	"context"
	"fmt"
	"log"
	"main/config"
	"sync"

	redis2 "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var RedisDb *redis2.Client
var MysqlDb *gorm.DB
var kafkaConn *kafka.Conn
var once sync.Once
var globalConfig *config.ParamsConfig

func Init(cfg *config.ParamsConfig) {
	globalConfig = cfg
}

func InitRedis(ctx context.Context) *redis2.Client {
	once.Do(func() {
		rdb := redis2.NewClient(&redis2.Options{
			Addr:          fmt.Sprintf("%s:%d", globalConfig.Redis.Host, globalConfig.Redis.Port),
			Protocol:      2,
			UnstableResp3: true,
			Password:      globalConfig.Redis.Password,
		})

		if _, err := rdb.Ping(ctx).Result(); err != nil {
			panic(fmt.Sprintf("Redis连接失败: %v", err))
		}
		RedisDb = rdb
	})
	return RedisDb
}

func InitMysql(ctx context.Context) *gorm.DB {
	if MysqlDb == nil {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			globalConfig.Mysql.User,
			globalConfig.Mysql.Password,
			globalConfig.Mysql.Host,
			globalConfig.Mysql.Port,
			globalConfig.Mysql.Database)
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
		topic := globalConfig.Kafka.Topic
		address := globalConfig.Kafka.Brokers
		partition := 0
		conn, err := kafka.DialLeader(context.Background(), "tcp", address, topic, partition)
		if err != nil {
			log.Fatal("failed to dial leader:", err)
		}
		return conn
	}
	return kafkaConn
}

func InitKafkaForConsumer(ctx context.Context) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{globalConfig.Kafka.Brokers},
		Topic:     globalConfig.Kafka.Topic,
		Partition: 0,
		MaxBytes:  10e6, // 10MB
		GroupID:   "my-first-kafka-consumer-group",
	})
	return r
}
