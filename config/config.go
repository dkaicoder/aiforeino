package config

import (
	"github.com/spf13/viper"
	"os"
	"sync"
)

type ParamsConfig struct {
	ApiKey    string `mapstructure:"APIKey"`
	Endpoint  string `mapstructure:"Endpoint"`
	Embedding string `mapstructure:"Embedding"`
	ChatModel string `mapstructure:"ChatModel"`
	Redis     struct {
		Host string `mapstructure:"Host"`
		Port int    `mapstructure:"Port"`
		DB   int    `mapstructure:"DB"`
	} `mapstructure:"Redis"`
}

var once sync.Once

var c *ParamsConfig

func Map() *ParamsConfig {
	once.Do(func() {
		v := viper.New()
		file, err := os.Open("config/config.yml")
		if err != nil {
			panic(err)
		}
		v.SetConfigType("yaml")
		err = v.ReadConfig(file)
		if err != nil {
			panic(err)
		}

		err = v.Unmarshal(&c)
		if err != nil {
			panic(err)
		}

		return
	})
	return c
}
