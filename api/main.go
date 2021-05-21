package main

import (
	"agones-minecraft/config"
	"agones-minecraft/db"
	"agones-minecraft/log"
	"agones-minecraft/routers"
	"agones-minecraft/services/auth"
)

func main() {
	config.LoadConfig()
	log.SetLog()
	r := routers.NewRouter()
	db.Init()
	auth.NewStore()
	port := config.GetPort()
	r.Run(":" + port)
}
