package sessions

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"agones-minecraft/config"
)

const (
	SessionName      = "agones-minecraft-api"
	StateCallbackKey = "state-callback"
	TokenKey         = "token"
)

var store cookie.Store

func NewStore() cookie.Store {
	gob.Register(oauth2.Token{})
	authKey, encKey := config.GetSessionSecret()
	store = cookie.NewStore(authKey, encKey)
	return store
}

func GetStore() cookie.Store {
	return store
}

func Sessions() gin.HandlerFunc {
	return sessions.Sessions(SessionName, store)
}

func NewState() (string, error) {
	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(tokenBytes[:]), nil
}

func AddStateFlash(c *gin.Context, state string) error {
	sess := sessions.Default(c)

	sess.AddFlash(state, StateCallbackKey)

	if err := sess.Save(); err != nil {
		zap.L().Error("error saving session", zap.Error(err))
		return err
	}
	return nil
}

func VerifyStateFlash(c *gin.Context, state string) (bool, error) {
	sess := sessions.Default(c)

	stateChallenge := sess.Flashes(StateCallbackKey)
	if err := sess.Save(); err != nil {
		zap.L().Warn("error saving session", zap.Error(err))
	}

	if state == "" {
		c.Status(http.StatusBadRequest)
		return false, fmt.Errorf("missing state")
	}

	if len(stateChallenge) < 1 {
		fmt.Println("missing state challenge")
		c.Status(http.StatusBadRequest)
		return false, fmt.Errorf("missing state challenge")
	}

	if state != stateChallenge[0] {
		c.Status(http.StatusUnauthorized)
		sess.Clear()
		return false, nil
	}

	return true, nil
}

func AddToken(c *gin.Context, tk *oauth2.Token) error {
	sess := sessions.Default(c)
	sess.Set(TokenKey, tk)
	if err := sess.Save(); err != nil {
		return err
	}
	return nil
}
