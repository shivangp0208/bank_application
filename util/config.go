package util

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBSource            string        `mapstructure:"dbSource"`
	DBDriver            string        `mapstructure:"dbDriver"`
	ServerAddress       string        `mapstructure:"serverAddress"`
	MinSecretKeyLength  int           `mapstructure:"minSecretKeyLength"`
	TokenSecretKey      string        `mapstructure:"tokenSecretKey"`
	TokenExpirationTime time.Duration `mapstructure:"tokenExpirationTime"`
}

func GetConfig() Config {
	config, err := LoadConfig("/home/shivangp0208/GoLang/Project/bank-application")
	if err != nil {
		log.Fatalf("unable to load configuration from config file: %v", err)
	}
	return config
}

func LoadConfig(path string) (config Config, err error) {

	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("unable to read the config: %v", err)
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to unmarshal the configuration: %v", err)
		return
	}

	return
}
