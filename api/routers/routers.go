package routers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agones-minecraft/config"
	v1Controllers "agones-minecraft/controllers/api/v1"
	"agones-minecraft/middleware/errors"
	"agones-minecraft/middleware/jwt"
	ginzap "agones-minecraft/middleware/log"
	"agones-minecraft/middleware/session"
	twitchMiddleware "agones-minecraft/middleware/twitch"
)

func NewRouter() *gin.Engine {
	if config.GetEnv() == config.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(zap.L(), false))

	engine.Use(func(c *gin.Context) {
		// enable CORS
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	engine.Use(session.Sessions())
	engine.Use(errors.HandleErrors())

	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := engine.Group("/api/v1")
	{
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
			game.POST("", v1Controllers.CreateGame)
			game.GET("/list", v1Controllers.ListGames)
			game.GET("/:name", v1Controllers.GetGame)
			game.DELETE("/:name", v1Controllers.DeleteGame)
		}
	}
	return engine
}
