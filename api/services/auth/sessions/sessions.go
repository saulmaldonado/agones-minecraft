package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

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

var Store cookie.Store

func Init() {
	Store = NewStore()
}

func NewStore() cookie.Store {
	authKey, encKey := config.GetSessionSecret()
	return cookie.NewStore(authKey, encKey)
}

func GetStore() cookie.Store {
	return Store
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

	// clear the session if its not new.
	sess.Clear()

	sess.AddFlash(state, StateCallbackKey)

	return sess.Save()
}

func VerifyStateFlash(c *gin.Context, state string) (bool, error) {
	sess := sessions.Default(c)

	stateChallenge := sess.Flashes(StateCallbackKey)
	if err := sess.Save(); err != nil {
		zap.L().Warn("error saving session", zap.Error(err))
	}

	if state == "" {
		return false, fmt.Errorf("missing state. try logging in again")
	}

	if len(stateChallenge) < 1 {
		return false, fmt.Errorf("missing state challenge. try logging in again")
	}

	if state != stateChallenge[0] {
		return false, fmt.Errorf("failed state challenge. try logging in again")
	}

	return true, nil
}

func AddToken(c *gin.Context, tk *oauth2.Token) error {
	sess := sessions.Default(c)
	sess.Set(TokenKey, tk)
	return sess.Save()
}
