package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"agones-minecraft/config"
)

const (
	SessionNamev1    = "agones-minecraft-api-twitch-oauth-v1"
	StateCallbackKey = "state-callback"
	TokenKey         = "token"
)

var (
	ErrMissingState          = errors.New("missing state from request")
	ErrMissingStateChallenge = errors.New("missing state from cookie")
	ErrFailedStateChallenge  = errors.New("failed state challenge")
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

	var err error

	stateChallenge := sess.Flashes(StateCallbackKey)
	if e := sess.Save(); e != nil {
		err = e
	}

	if state == "" {
		return false, ErrMissingState
	}

	if len(stateChallenge) < 1 {
		return false, ErrMissingStateChallenge
	}

	if state != stateChallenge[0] {
		return false, ErrFailedStateChallenge
	}

	return true, err
}
