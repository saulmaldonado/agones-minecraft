package v1Controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	v1Err "agones-minecraft/errors/v1"
	"agones-minecraft/middleware/jwt"
	"agones-minecraft/middleware/twitch"
	userv1Model "agones-minecraft/models/v1/user"
	apiErr "agones-minecraft/resource/api/v1/errors"
	userv1Resource "agones-minecraft/resource/api/v1/user"
	userv1Service "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/mc"
)

var (
	ErrMissingUserId error = errors.New("user id reference missing from request")
	ErrUserNotFound  error = errors.New("user not found")
)

func GetMe(c *gin.Context) {
	v, ok := c.Get(twitch.SubjectKey)
	userId, okUserId := v.(string)
	if !ok || !okUserId {
		c.Error(apiErr.NewUnauthorizedError(ErrMissingUserId, v1Err.ErrMissingUserId))
		return
	}

	var user userv1Model.User
	if err := userv1Service.GetUserById(uuid.MustParse(userId), &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(apiErr.NewNotFoundError(ErrUserNotFound, v1Err.ErrUserNotFound))
			return
		}
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingUser))
		return
	}

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
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			c.Errors = append(c.Errors, apiErr.NewValidationError(verrs, v1Err.ErrEditUserValidation)...)
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrMalformedJSON))
		}
		return
	}

	mcUser, err := mc.GetUser(body.MCUsername)
	if err != nil {
		if e, ok := err.(*mc.ErrMcUserNotFound); ok {
			c.Error(apiErr.NewNotFoundError(e, v1Err.ErrMcUserNotFound))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingMcUser))
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
			c.Error(apiErr.NewNotFoundError(ErrUserNotFound, v1Err.ErrUserNotFound))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrUpdatingUser))
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
