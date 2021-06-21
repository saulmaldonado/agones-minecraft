package sessions

import (
	"agones-minecraft/config"
	"log"

	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
)

var Store redis.Store

func Init() {
	var err error
	Store, err = NewStore()
	if err != nil {
		log.Fatal(err)
	}
}

func NewStore() (redis.Store, error) {
	authKey, encKey := config.GetSessionSecret()
	redisConfig := config.GetRedisCreds()

	store, err := redis.NewStore(10, "tcp", redisConfig.Address, redisConfig.Password, authKey, encKey)
	if err != nil {
		log.Fatal(err)
	}

	return store, nil
}

func GetStore() cookie.Store {
	return Store
}
