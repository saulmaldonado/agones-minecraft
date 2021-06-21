package config

import (
	"log"

	"github.com/spf13/viper"
)

// env variables
const (
	ENV                  = "ENV"
	PORT                 = "PORT"
	DB_USER              = "DB_USER"
	DB_PASSWORD          = "DB_PASSWORD"
	DB_HOST              = "DB_HOST"
	DB_PORT              = "DB_PORT"
	DB_NAME              = "DB_NAME"
	OAUTH_SESSION_SECRET = "OAUTH_SESSION_SECRET"
	SESSION_SECRET       = "SESSION_SECRET"
	TWITCH_CLIENT_ID     = "TWITCH_CLIENT_ID"
	TWITCH_CLIENT_SECRET = "TWITCH_CLIENT_SECRET"
	TWITCH_REDIRECT      = "TWITCH_REDIRECT"
	REDIS_ADDRESS        = "REDIS_ADDRESS"
	REDIS_PASSWORD       = "REDIS_PASSWORD"
	JWT_SECRET           = "JWT_SECRET"
	DNS_ZONE             = "DNS_ZONE"
)

const (
	Production  environment = "production"
	Development environment = "development"
)

// Loads App config from .env file and environment variables with the latter taking precedence
func InitConfig() {
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")

	viper.SetDefault(ENV, Development) // default to development
	viper.SetDefault(PORT, "8080")     // default to 8080

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal(err)
		}
	}
}

type environment string

type DBConfig struct {
	Password string
	Hostname string
	Port     string
	Name     string
	User     string
}

func GetDBConfig() *DBConfig {
	return &DBConfig{
		Password: viper.GetString(DB_PASSWORD),
		Hostname: viper.GetString(DB_HOST),
		Port:     viper.GetString(DB_PORT),
		Name:     viper.GetString(DB_NAME),
		User:     viper.GetString(DB_USER),
	}
}

// Returns current environment
func GetEnv() environment {
	v := viper.GetString(ENV)

	if v != string(Development) && v != string(Production) {
		log.Fatal("invalid ENV value")
	}

	return environment(v)
}

// Returns PORT from config
func GetPort() string {
	return viper.GetString(PORT)
}

// Returns secrets for oauth sessions
func GetOAuthSessionSecret() (authKey []byte, encKey []byte) {
	keys := viper.GetStringSlice(OAUTH_SESSION_SECRET)
	if len(keys) < 2 {
		log.Fatal("missing SESSION_SECRET authenticationa and encryption keys")
	}

	return []byte(keys[0]), []byte(keys[1])
}

// Returns secrets for sessions
func GetSessionSecret() (authKey []byte, encKey []byte) {
	keys := viper.GetStringSlice(SESSION_SECRET)
	if len(keys) < 2 {
		log.Fatal("missing SESSION_SECRET authenticationa and encryption keys")
	}

	return []byte(keys[0]), []byte(keys[1])
}

// ID, secret and current redirect for Twitch
type TwitchConfig struct {
	ClientID     string
	ClientSecret string
	Redirect     string
}

// Returns Twitch configuration
func GetTwichCreds() *TwitchConfig {
	return &TwitchConfig{
		ClientID:     viper.GetString(TWITCH_CLIENT_ID),
		ClientSecret: viper.GetString(TWITCH_CLIENT_SECRET),
		Redirect:     viper.GetString(TWITCH_REDIRECT),
	}
}

// Returns JWT HS256 secret
func GetJWTSecret() string {
	return viper.GetString(JWT_SECRET)
}

// Redis address and password
type RedisConfig struct {
	Address  string
	Password string
}

// Returns redis configuration
func GetRedisCreds() *RedisConfig {
	return &RedisConfig{
		Address:  viper.GetString(REDIS_ADDRESS),
		Password: viper.GetString(REDIS_PASSWORD),
	}
}

// Returns DNS zone
func GetDNSZone() string {
	return viper.GetString(DNS_ZONE)
}
