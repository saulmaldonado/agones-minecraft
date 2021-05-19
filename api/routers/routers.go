package routers

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agones-minecraft/config"
	v1Controllers "agones-minecraft/controllers/api/v1"
)

func NewRouter() *gin.Engine {
	if config.GetEnv() == config.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(zap.L(), true))

	engine.Use(func(c *gin.Context) {
		// enable CORS
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/ping", v1Controllers.Ping)
	}

	return engine
}
