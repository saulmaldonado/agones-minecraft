package config

import (
	"log"

	"github.com/spf13/viper"
)

const (
	ENV                  = "ENV"
	PORT                 = "PORT"
	DB_USER              = "DB_USER"
	DB_PASSWORD          = "DB_PASSWORD"
	DB_HOST              = "DB_HOST"
	DB_PORT              = "DB_PORT"
	DB_NAME              = "DB_NAME"
	SESSION_SECRET       = "SESSION_SECRET"
	TWITCH_CLIENT_ID     = "TWITCH_CLIENT_ID"
	TWITCH_CLIENT_SECRET = "TWITCH_CLIENT_SECRET"
	TWITCH_REDIRECT      = "TWITCH_REDIRECT"
	REDIS_ADDRESS        = "REDIS_ADDRESS"
	REDIS_PASSWORD       = "REDIS_PASSWORD"
	JWT_SECRET           = "JWT_SECRET"

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

// Returns PORT from config
func GetPort() string {
	return viper.GetString(PORT)
}

// Returns DB_USER, DB_PASSWORD, DB_HOST, DB_PORT and DB_NAME from config
func GetDB() (string, string, string, string, string) {
	return viper.GetString(DB_USER),
		viper.GetString(DB_PASSWORD),
		viper.GetString(DB_HOST),
		viper.GetString(DB_PORT),
		viper.GetString(DB_NAME)
}

func GetSessionSecret() (authKey []byte, encKey []byte) {
	keys := viper.GetStringSlice(SESSION_SECRET)
	if len(keys) < 2 {
		log.Fatal("missing SESSION_SECRET authenticationa and encryption keys")
	}
	authKey = []byte(keys[0])
	encKey = []byte(keys[1])
	return
}

func GetTwichCreds() (clientId string, clientSecret string, redirect string) {
	return viper.GetString(TWITCH_CLIENT_ID), viper.GetString(TWITCH_CLIENT_SECRET), viper.GetString(TWITCH_REDIRECT)
}

func GetJWTSecret() string {
	return viper.GetString(JWT_SECRET)
}

func GetRedisCreds() (address string, password string) {
	return viper.GetString(REDIS_ADDRESS), viper.GetString(REDIS_PASSWORD)
}
