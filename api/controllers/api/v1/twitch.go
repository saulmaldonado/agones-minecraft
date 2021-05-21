package v1Controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	sessionsauth "agones-minecraft/services/auth/sessions"
	twitchauth "agones-minecraft/services/auth/twitch"
)

func TwitchLogin(c *gin.Context) {
	config := twitchauth.NewTwitchConfig()

	state, err := sessionsauth.NewState()
	if err != nil {
		zap.L().Error("error generating new state", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := sessionsauth.AddStateFlash(c, state); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	http.Redirect(c.Writer, c.Request, config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func TwitchCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	ok, err := sessionsauth.VerifyStateFlash(c, state)
	if err != nil {
		c.Status(http.StatusBadRequest)
		zap.L().Error("error verifying state", zap.Error(err))
		return
	}

	if !ok {
		c.Status(http.StatusUnauthorized)
		zap.L().Warn("failed state challenge")
		return
	}

	config := twitchauth.NewTwitchConfig()

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		zap.L().Error("error exchaning code for token", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, token)
}
