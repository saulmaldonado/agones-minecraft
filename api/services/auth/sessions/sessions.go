package sessions

import (
	"agones-minecraft/config"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const (
	SessionName = "agones-minecraft-api"
)

var store cookie.Store

func NewStore() cookie.Store {
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
