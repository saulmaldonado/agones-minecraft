package main

import (
	"agones-minecraft/config"
	"agones-minecraft/log"
	"agones-minecraft/routers"
)

func main() {
	config.LoadConfig()
	log.SetLog()
	r := routers.NewRouter()

	port := config.GetPort()
	r.Run(":" + port)
}
