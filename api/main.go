package main

import (
	"agones-minecraft/config"
	"agones-minecraft/db"
	"agones-minecraft/log"
	"agones-minecraft/routers"
	"agones-minecraft/services/auth/jwt"
	"agones-minecraft/services/auth/sessions"
	"agones-minecraft/services/auth/twitch"
	"agones-minecraft/services/validator"
)

func main() {
	config.LoadConfig()
	log.SetLog()

	sessions.NewStore()
	db.Init()

	twitch.NewODICProvider()
	jwt.New()
	validator.InitV1()

	r := routers.NewRouter()

	port := config.GetPort()
	r.Run(":" + port)
}
