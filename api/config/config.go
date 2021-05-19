package config

import (
	"log"

	"github.com/spf13/viper"
)

const (
	ENV         = "ENV"
	PORT        = "PORT"
	Production  = "production"
	Development = "development"
)

// Loads App config from .env file and environment variables with the latter taking precedence
func LoadConfig() {
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")
	viper.SetDefault(ENV, Development)
	viper.SetDefault(PORT, "8080")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal(err)
		}
	}

}

// Returns ENV from config
func GetEnv() string {
	return viper.GetString(ENV)
}

func GetPort() string {
	return viper.GetString(PORT)
}
