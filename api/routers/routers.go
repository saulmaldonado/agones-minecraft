package routers

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agones-minecraft/config"
	v1Controllers "agones-minecraft/controllers/api/v1"
	"agones-minecraft/services/auth/sessions"
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

	engine.Use(sessions.Sessions())

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
	}
	return engine
}
