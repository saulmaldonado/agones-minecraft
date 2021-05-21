package v1Controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	sessionsauth "agones-minecraft/services/auth/sessions"
	twitchauth "agones-minecraft/services/auth/twitch"
)

const (
	StateCallbackKey = "state-callback"
)

func TwitchLogin(c *gin.Context) {
	sess := sessions.Default(c)

	state, err := sessionsauth.NewState()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		zap.L().Error("error generating new state", zap.Error(err))
		return
	}

	sess.AddFlash(state, StateCallbackKey)

	if err := sess.Save(); err != nil {
		c.Status(http.StatusInternalServerError)
		zap.L().Error("error saving session", zap.Error(err))
		return
	}

	config := twitchauth.NewTwitchConfig()

	http.Redirect(c.Writer, c.Request, config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func TwitchCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	sess := sessions.Default(c)

	stateChallenge := sess.Flashes(StateCallbackKey)
	if state == "" {
		log.Println("missing state")
		c.Status(http.StatusBadRequest)
		return
	}

	if len(stateChallenge) < 1 {
		fmt.Println("missing state challenge")
		c.Status(http.StatusBadRequest)
		return
	}

	if state != stateChallenge[0] {
		zap.L().Warn("non-matching states", zap.String("state", state), zap.String("stateChallenge", stateChallenge[0].(string)))
		c.Status(http.StatusUnauthorized)
		sess.Clear()
		return
	}

	config := twitchauth.NewTwitchConfig()

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		zap.L().Error("error exchaning code for token", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := sess.Save(); err != nil {
		c.Status(http.StatusInternalServerError)
		zap.L().Error("error saving session", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, token)
}
