package main

import (
	"agones-minecraft/config"
	"agones-minecraft/db"
	"agones-minecraft/log"
	"agones-minecraft/routers"
	"agones-minecraft/services/auth/sessions"
	twitch "agones-minecraft/services/auth/twitch"
)

func main() {
	config.LoadConfig()
	log.SetLog()

	sessions.NewStore()
	db.Init()

	twitch.NewODICProvider()

	r := routers.NewRouter()

	port := config.GetPort()
	r.Run(":" + port)
}
