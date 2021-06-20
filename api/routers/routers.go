package routers

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agones-minecraft/config"
	v1Controllers "agones-minecraft/controllers/api/v1"
	"agones-minecraft/db"
	apiErr "agones-minecraft/middleware/errors"
	"agones-minecraft/middleware/jwt"
	ginzap "agones-minecraft/middleware/log"
	"agones-minecraft/middleware/session"
	jwtServicev1 "agones-minecraft/services/auth/jwt"
	"agones-minecraft/services/k8s/agones"

	twitchMiddleware "agones-minecraft/middleware/twitch"
)

func NewRouter() *gin.Engine {
	if config.GetEnv() == config.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(zap.L(), false))

	engine.GET("/health", func(c *gin.Context) {
		var database bool
		var memorystore bool
		var cluster bool

		if err := db.Ping(); err == nil {
			database = true
		}

		if err := jwtServicev1.Get().Ping(); err == nil {
			memorystore = true
		}

		if err := agones.Client().Ping(); err == nil {
			cluster = true
		}

		c.JSON(http.StatusOK, gin.H{
			"api":         true,
			"database":    database,
			"memorystore": memorystore,
			"cluster":     cluster,
		})
	})

	AddV1Router(engine)

	return engine
}

func AddV1Router(r *gin.Engine) {

	v1 := r.Group("/api/v1")

	// Allow all origins
	v1.Use(cors.Default())
	v1.Use(session.Sessions())
	v1.Use(apiErr.HandleErrors())

	twitch := v1.Group("/twitch")
	{
		twitch.GET("/login", v1Controllers.TwitchLogin)
		twitch.GET("/callback", v1Controllers.TwitchCallback)
	}

	auth := v1.Group("/auth")
	{
		auth.POST("/refresh", v1Controllers.Refresh)
		auth.Use(jwt.Authorizer())
		auth.POST("/logout", v1Controllers.Logout)
	}

	user := v1.Group("/user")
	{
		user.Use(jwt.Authorizer())
		user.Use(twitchMiddleware.Authorizer())
		user.GET("/me", v1Controllers.GetMe)
		user.POST("/me", v1Controllers.EditMe)
	}

	game := v1.Group("/game")
	{
		game.GET("", v1Controllers.ListGames)
		game.GET("/:name/status", jwt.Authenticator(), v1Controllers.GetGameState)
		game.GET("/:name", v1Controllers.GetGame)

		game.Use(jwt.Authorizer())
		game.POST("/java", v1Controllers.CreateJava)
		game.POST("/bedrock", v1Controllers.CreateBedrock)
		game.DELETE("/:name", v1Controllers.DeleteGame)
	}
}
