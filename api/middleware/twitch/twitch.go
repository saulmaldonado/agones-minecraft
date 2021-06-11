package twitch

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	v1Err "agones-minecraft/errors/v1"
	apiErr "agones-minecraft/resource/api/v1/errors"
	userv1Service "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/auth/twitch"
)

const (
	SubjectKey string = "JWT_SUBJECT"
	TokenIDKey string = "JWT_ID"
	UserKey    string = "USER"
)

var (
	ErrTwitchCredentialsNotFound error = errors.New("user's Twitch credentials not found or have been deleted. login to renew crednentials")
)

func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		v := c.GetString(SubjectKey)
		userId := uuid.MustParse(v)

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
