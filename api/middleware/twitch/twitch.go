package twitch

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"agones-minecraft/config"
	twitchv1Model "agones-minecraft/models/v1/twitch"
	"agones-minecraft/resource/api/v1/errors"
	userv1Service "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/auth/jwt"
	"agones-minecraft/services/auth/twitch"
)

const (
	SubjectKey string = "JWT_SUBJECT"
	TokenIDKey string = "JWT_ID"
	UserKey    string = "USER"
)

func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetString(SubjectKey)
		var twitchTokens twitchv1Model.TwitchToken
		// Gets tokens for user from DB
		if err := userv1Service.GetUserTwitchTokens(uuid.MustParse(userId), &twitchTokens); err != nil {
			if err == gorm.ErrRecordNotFound {
				c.Errors = append(c.Errors, errors.NewUnauthorizedError(fmt.Errorf("user's Twitch credentials not found or have been deleted. login again to renew credentials and tokens")))
			} else {
				c.Errors = append(c.Errors, errors.NewInternalServerError(err))
			}
			c.Abort()
			return
		}

		// Validate found access tokens with Twitch
		if err := twitch.ValidateToken(*twitchTokens.TwitchAccessToken); err != nil {
			// If access token is invalid attempt refresh
			if err == twitch.ErrInvalidAccessToken {
				clientId, clientSecret, _ := config.GetTwichCreds()
				// Refresh tokens with Twitch
				token, err := twitch.Refresh(*twitchTokens.TwitchRefreshToken, clientId, clientSecret)
				if err != nil {
					// Both tokens are invalid. New login is required
					if err == twitch.ErrInvalidatedTokens {
						c.Errors = append(c.Errors, errors.NewUnauthorizedError(err))
						// revoke api tokens to force new login
						tokenId := c.GetString(TokenIDKey)
						if err := jwt.Get().Delete(tokenId); err != nil {
							zap.L().Warn("error revoking app tokens", zap.Error(err))
						}
					} else {
						c.Errors = append(c.Errors, errors.NewInternalServerError(err))
					}
					c.Abort()
					return
				}

				twitchTokens.TwitchAccessToken = &token.AccessToken
				twitchTokens.TwitchRefreshToken = &token.RefreshToken

				// Update tokens for user in database
				if err := userv1Service.UpdateUserTwitchTokens(uuid.MustParse(userId), &twitchTokens); err != nil {
					c.Errors = append(c.Errors, errors.NewInternalServerError(err))
					c.Abort()
					return
				}
			} else {
				c.Errors = append(c.Errors, errors.NewInternalServerError(err))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
