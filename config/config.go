package config

import (
	"sync"

	"github.com/spf13/viper"
)

type ParamsConfig struct {
	ApiKey     string `mapstructure:"APIKey"`
	ExportHost string `mapstructure:"ExportHost"`
	ChatModel  string `mapstructure:"ChatModel"`
	Embedding  struct {
		ApiKey string `mapstructure:"APIKey"`
	} `mapstructure:"Embedding"`
	Redis struct {
		Host     string `mapstructure:"Host"`
		Port     int    `mapstructure:"Port"`
		DB       int    `mapstructure:"DB"`
		Password string `mapstructure:"Password"`
	} `mapstructure:"Redis"`
	Mysql struct {
		Host     string `mapstructure:"Host"`
		Port     int    `mapstructure:"Port"`
		User     string `mapstructure:"User"`
		Password string `mapstructure:"Password"`
		Database string `mapstructure:"Database"`
	} `mapstructure:"Mysql"`
	Kafka struct {
		Brokers  string `mapstructure:"Brokers"`
		Topic    string `mapstructure:"Topic"`
		Username string `mapstructure:"Username"`
		Password string `mapstructure:"Password"`
	} `mapstructure:"Kafka"`
	Langfuse struct {
		Host      string `mapstructure:"Host"`
		PublicKey string `mapstructure:"PublicKey"`
		SecretKey string `mapstructure:"SecretKey"`
	} `mapstructure:"Langfuse"`
}

var once sync.Once

var C *ParamsConfig

func InitConfig() *ParamsConfig {
	once.Do(func() {
		v := viper.New()
		v.AddConfigPath("./config")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		err := v.ReadInConfig()
		if err != nil {
			panic(err)
		}
		err = v.Unmarshal(&C)
		if err != nil {
			panic(err)
		}

		return
	})
	return C
}

func GetConfig() *ParamsConfig {
	return C
}
