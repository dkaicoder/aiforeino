package config

import (
	"sync"

	"github.com/spf13/viper"
)

type ParamsConfig struct {
	ApiKey    string `mapstructure:"APIKey"`
	Embedding string `mapstructure:"Embedding"`
	ChatModel string `mapstructure:"ChatModel"`
	Redis     struct {
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
		Brokers string `mapstructure:"Brokers"`
		Topic   string `mapstructure:"Topic"`
	} `mapstructure:"Kafka"`
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
