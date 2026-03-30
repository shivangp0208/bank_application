package util

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DBSource           string `mapstructure:"dbSource"`
	DBDriver           string `mapstructure:"dbDriver"`
	ServerAddress      string `mapstructure:"serverAddress"`
	MinSecretKeyLength int    `mapstructure:"minSecretKeyLength"`
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
