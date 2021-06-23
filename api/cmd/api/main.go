package api

import (
	"agones-minecraft/config"
	"agones-minecraft/db"
	"agones-minecraft/log"
	"agones-minecraft/routers"
	sessions "agones-minecraft/services/auth/sessions"
	"agones-minecraft/services/auth/twitch"
	appHttp "agones-minecraft/services/http"
	"agones-minecraft/services/k8s"
	"agones-minecraft/services/k8s/agones"
	"agones-minecraft/services/validator"
)

func Run() error {
	// Load environment variables and .env config
	config.InitConfig()
	// Sets global zap logger
	log.Init()
	// Initializes k8s cluster config
	k8s.InitConfig()
	// Initializes app http client
	appHttp.Init()
	// Connects to k8s cluster and initializes agones client and informer
	agones.Init()
	// Initializes session and oauth session redis store
	sessions.Init()
	// Initializes database connections and migrates (in development)
	db.Init()
	// Initializes Twitch ODIC provider for login with Twitch
	twitch.Init()
	// Initializes custom validators
	validator.InitV1()

	r := routers.NewRouter()

	port := config.GetPort()
	return r.Run(":" + port)
}
