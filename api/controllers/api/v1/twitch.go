package v1Controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agones-minecraft/services/auth"
)

const (
	StateSessionKey  = "state"
	StateCallbackKey = "state-callback"
)

func TwitchLogin(c *gin.Context) {
	store := auth.GetStore()
	sess, err := store.Get(c.Request, StateSessionKey)
	if err != nil {
		zap.L().Warn("error decoding state session", zap.Error(err))
		err = nil
	}

	state, err := auth.NewState()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		zap.L().Error("error generating new state", zap.String("sessionId", sess.ID))
		return
	}

	sess.AddFlash(state, StateCallbackKey)

	if err := sess.Save(c.Request, c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
		zap.L().Error("error saving session", zap.Error(err))
		return
	}

	config := auth.NewTwitchConfig()

	http.Redirect(c.Writer, c.Request, config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func TwitchCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	store := auth.GetStore()
	sess, err := store.Get(c.Request, StateSessionKey)
	if err != nil {
		zap.L().Warn("error decoding state session")
		err = nil
	}

	stateChallenge := sess.Flashes(StateCallbackKey)
	if state == "" {
		fmt.Println("missing state")
		c.Status(http.StatusBadRequest)
		return
	}

	if len(stateChallenge) < 1 {
		fmt.Println("missing state challenge")
		c.Status(http.StatusBadRequest)
		return
	}

	if state != stateChallenge[0] {
		c.Status(http.StatusUnauthorized)
	}

	config := auth.NewTwitchConfig()

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		zap.L().Error("error exchaning code for token", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, token)
}
