package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	SessionNamev1    string = "agones-minecraft-api-twitch-oauth-v1"
	StateCallbackKey string = "state-callback"

	SessionPath string = "/api/v1"
	// 5 min
	SessionMaxAge int = 300
)

var (
	ErrMissingState          = errors.New("missing state from request")
	ErrMissingStateChallenge = errors.New("missing state from cookie")
	ErrFailedStateChallenge  = errors.New("failed state challenge")
)

func NewState() (string, error) {
	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(tokenBytes[:]), nil
}

func AddStateFlash(c *gin.Context, state string) error {
	sess := sessions.DefaultMany(c, SessionNamev1)

	// clear the session if its not new.
	sess.Clear()

	sess.AddFlash(state, StateCallbackKey)

	sess.Options(sessions.Options{
		Path:     SessionPath,
		MaxAge:   SessionMaxAge,
		Secure:   true,
		HttpOnly: true,
	})

	return sess.Save()
}

func VerifyStateFlash(c *gin.Context, state string) (bool, error) {
	sess := sessions.DefaultMany(c, SessionNamev1)

	var err error

	stateChallenge := sess.Flashes(StateCallbackKey)
	if e := DestorySession(sess); e != nil {
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

func DestorySession(sess sessions.Session) error {
	sess.Options(sessions.Options{
		Path:     SessionPath,
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
	})

	return sess.Save()
}
