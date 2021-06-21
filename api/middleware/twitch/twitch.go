package twitch

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	v1Err "agones-minecraft/errors/v1"
	sessionsMiddleware "agones-minecraft/middleware/session"
	apiErr "agones-minecraft/resources/api/v1/errors"
	userv1Service "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/auth/sessions"
	"agones-minecraft/services/auth/twitch"
)

const (
	TokenIDKey string = "JWT_ID"
	UserKey    string = "USER"
)

var (
	ErrTwitchCredentialsNotFound error = errors.New("user's Twitch credentials not found or have been deleted. login to renew crednentials")
)

// Validates and refreshes users stored OAuth Twitch tokens with Twitch OAuth servers
// Returns an unauthorized error if tokens have been invalited by Twitch
// Middleware gets request userId context from sessions
func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := sessions.GetSessionUserId(c)
		if userId == uuid.Nil {
			c.Error(apiErr.NewUnauthorizedError(sessionsMiddleware.ErrUnauthorizedSession, v1Err.ErrUnautorizedSession))
			c.Abort()
			return
		}

		if err := userv1Service.ValidateAndRefreshTwitchTokensForUser(userId); err != nil {
			switch err {
			case gorm.ErrRecordNotFound:
				c.Error(apiErr.NewUnauthorizedError(ErrTwitchCredentialsNotFound, v1Err.ErrTwitchCredentialsNotFound))
			case twitch.ErrTwitchCredentialsInvalid:
				c.Error(apiErr.NewUnauthorizedError(twitch.ErrTwitchCredentialsInvalid, v1Err.ErrTwitchCredentialsInvalid))
			default:
				c.Error(apiErr.NewInternalServerError(err, v1Err.ErrValidatingTwitchToken))
			}
			c.Abort()
			return
		}

		c.Next()
	}
}
