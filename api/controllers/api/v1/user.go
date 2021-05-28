package v1Controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"agones-minecraft/config"
	"agones-minecraft/middleware/jwt"
	"agones-minecraft/models"
	"agones-minecraft/resource/api/v1/errors"
	userv1Resource "agones-minecraft/resource/api/v1/user"
	userv1Service "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/auth/twitch"
	"agones-minecraft/services/mc"
)

func GetMe(c *gin.Context) {
	userId := c.GetString(jwt.ContextKey)
	var user models.User
	if err := userv1Service.GetUserById(userId, &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Errors = append(c.Errors, errors.NewNotFoundError(err))
		} else {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		}
		return
	}

	accessToken := user.TwitchToken.TwitchAccessToken

	if err := twitch.ValidateToken(*accessToken); err != nil {
		if err == twitch.ErrInvalidAccessToken {
			zap.L().Info("invalid Twitch access token", zap.String("user", user.ID.String()))
			clientId, clientSecret, _ := config.GetTwichCreds()
			token, err := twitch.Refresh(*user.TwitchToken.TwitchRefreshToken, clientId, clientSecret)

			if err != nil {
				if err == twitch.ErrInvalidatedTokens {
					c.Errors = append(c.Errors, errors.NewUnauthorizedError(err))
				} else {
					c.Errors = append(c.Errors, errors.NewInternalServerError(err))
				}
				return
			}

			user = models.User{
				ID: uuid.MustParse(userId),
				TwitchToken: models.TwitchToken{
					TwitchAccessToken:  &token.AccessToken,
					TwitchRefreshToken: &token.RefreshToken,
				},
			}

			if err := userv1Service.EditUser(&user); err != nil {
				c.Errors = append(c.Errors, errors.NewInternalServerError(err))
				return
			}
			zap.L().Info("refreshed Twitch access and refresh token", zap.String("user", user.ID.String()))
		} else {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
			zap.L().Warn("error validating user with Twitch", zap.Error(err))
			return
		}
	}

	fmt.Println(user)

	foundUser := userv1Resource.User{
		ID:             user.ID,
		Email:          *user.Email,
		EmailVerified:  *user.EmailVerified,
		TwitchID:       user.TwitchID,
		TwitchUsername: user.TwitchUsername,
		MCUsername:     user.MCUsername,
		MCUUID:         user.MCUUID,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	c.JSON(http.StatusOK, foundUser)
}

func EditMe(c *gin.Context) {
	userId := c.GetString(jwt.ContextKey)

	var body userv1Resource.EditUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Errors = append(c.Errors, errors.NewBadRequestError(err))
		return
	}

	mcUser, err := mc.GetUser(body.MCUsername)
	if err != nil {
		if e, ok := err.(*mc.ErrMcUserNotFound); ok {
			c.Errors = append(c.Errors, errors.NewNotFoundError(e))
		} else {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		}
		return
	}

	user := models.User{
		ID:         uuid.MustParse(userId),
		MCUsername: &mcUser.Username,
		MCUUID:     &mcUser.UUID,
	}

	if err := userv1Service.EditUser(&user); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Errors = append(c.Errors, errors.NewNotFoundError(fmt.Errorf("userId %s not found", userId)))
		} else {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		}
		return
	}

	updatedUser := userv1Resource.User{
		ID:             user.ID,
		Email:          *user.Email,
		EmailVerified:  *user.EmailVerified,
		TwitchUsername: user.TwitchUsername,
		TwitchID:       user.TwitchID,
		MCUsername:     user.MCUsername,
		MCUUID:         user.MCUUID,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	c.JSON(http.StatusOK, updatedUser)
}
