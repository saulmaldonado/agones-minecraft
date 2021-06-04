package v1Controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"agones-minecraft/middleware/jwt"
	"agones-minecraft/middleware/twitch"
	userv1Model "agones-minecraft/models/v1/user"
	"agones-minecraft/resource/api/v1/errors"
	userv1Resource "agones-minecraft/resource/api/v1/user"
	userv1Service "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/mc"
)

func GetMe(c *gin.Context) {
	v, ok := c.Get(twitch.SubjectKey)
	userId, okUserId := v.(string)
	if !ok || !okUserId {
		c.Errors = append(c.Errors, errors.NewNotFoundError(fmt.Errorf("user not found")))
		return
	}

	var user userv1Model.User
	userv1Service.GetUserById(uuid.MustParse(userId), &user)

	foundUser := userv1Resource.User{
		ID:             user.ID,
		Email:          *user.Email,
		EmailVerified:  *user.EmailVerified,
		TwitchID:       user.TwitchID,
		TwitchUsername: user.TwitchUsername,
		TwitchPicture:  user.TwitchPicture,
		MCUsername:     user.MCUsername,
		MCUUID:         user.MCUUID,
		LastLogin:      user.LastLogin,
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

	user := userv1Model.User{
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
		TwitchID:       user.TwitchID,
		TwitchUsername: user.TwitchUsername,
		TwitchPicture:  user.TwitchPicture,
		MCUsername:     user.MCUsername,
		MCUUID:         user.MCUUID,
		LastLogin:      user.LastLogin,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	c.JSON(http.StatusOK, updatedUser)
}
