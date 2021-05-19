package main

import (
	"agones-minecraft/config"
	"agones-minecraft/db"
	"agones-minecraft/log"
	"agones-minecraft/routers"
)

func main() {
	config.LoadConfig()
	log.SetLog()
	r := routers.NewRouter()
	db.Init()
	port := config.GetPort()
	r.Run(":" + port)
}
