package session

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	s "github.com/gin-contrib/sessions"

	v1Err "agones-minecraft/errors/v1"
	apiErr "agones-minecraft/resources/api/v1/errors"
	"agones-minecraft/services/auth/sessions"
	oauthSessions "agones-minecraft/services/auth/sessions/oauth"
)

var (
	ErrUnauthorizedSession error = errors.New("missing or invalid sesssion. log in again")
)

const (
	SessionUserIDKey string = "SESSION_USER_ID"
)

func Sessions() gin.HandlerFunc {
	return s.SessionsMany(
		[]string{sessions.SessionNamev1, oauthSessions.SessionNamev1},
		sessions.GetStore(),
	)
}

func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := sessions.GetSessionUserId(c)
		if id == uuid.Nil {
			c.Error(apiErr.NewUnauthorizedError(ErrUnauthorizedSession, v1Err.ErrUnautorizedSession))
			c.Abort()
			return
		}

		c.Set(SessionUserIDKey, id)
		c.Next()
	}
}

func Authenticator() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := sessions.GetSessionUserId(c)
		if id != uuid.Nil {
			c.Set(SessionUserIDKey, id)
		}

		c.Next()
	}
}
