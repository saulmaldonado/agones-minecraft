package v1Controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"agones-minecraft/middleware/jwt"
	"agones-minecraft/models"
	"agones-minecraft/resource/api/v1/errors"
	userv1Resource "agones-minecraft/resource/api/v1/user"
	userv1Service "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/auth/twitch"
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

	accessToken := user.TwitchAccessToken

	if err := twitch.ValidateToken(*accessToken); err != nil {
		c.Errors = append(c.Errors, errors.NewNotFoundError(err))
		zap.L().Warn("error validating user with Twitch", zap.Error(err))
		return
	}

	foundUser := userv1Resource.User{
		ID:             user.ID,
		Email:          user.Email,
		TwitchUsername: &user.TwitchUsername,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	c.JSON(http.StatusOK, foundUser)
}
