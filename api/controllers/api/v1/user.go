package v1Controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	v1Err "agones-minecraft/errors/v1"
	"agones-minecraft/middleware/jwt"
	"agones-minecraft/middleware/twitch"
	mcv1Model "agones-minecraft/models/v1/mc"
	"agones-minecraft/models/v1/model"
	userv1Model "agones-minecraft/models/v1/user"
	apiErr "agones-minecraft/resources/api/v1/errors"
	userv1Resource "agones-minecraft/resources/api/v1/user"
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
	if err := userv1Service.GetUserById(&user, uuid.MustParse(userId)); err != nil {
		if err == pg.ErrNoRows {
			c.Error(apiErr.NewNotFoundError(ErrUserNotFound, v1Err.ErrUserNotFound))
			return
		}
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingUser))
		return
	}

	foundUser := userv1Resource.User{
		ID:            user.ID,
		Email:         user.TwitchAccount.Email,
		EmailVerified: user.TwitchAccount.EmailVerified,
		TwitchAccount: userv1Resource.TwitchAccount{
			TwitchID:       user.TwitchAccount.ID,
			TwitchUsername: user.TwitchAccount.Username,
			TwitchPicture:  user.TwitchAccount.Picture,
		},
		MCAccount: userv1Resource.MCAccount{
			MCUsername: user.MCAccount.Username,
			MCUUID:     user.MCAccount.ID,
		},
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, foundUser)
}

func EditMe(c *gin.Context) {
	userId := uuid.MustParse(c.GetString(jwt.ContextKey))

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
		Model: model.Model{
			ID: userId,
		},
		MCAccount: &mcv1Model.MCAccount{
			Model:    model.Model{ID: mcUser.UUID},
			Username: mcUser.Username,
		},
	}

	if err := userv1Service.EditMCAccount(user.MCAccount, user.ID); err != nil {
		if err == pg.ErrNoRows {
			c.Error(apiErr.NewNotFoundError(ErrUserNotFound, v1Err.ErrUserNotFound))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrUpdatingUser))
		}
		return
	}

	updatedUser := userv1Resource.User{
		ID:            user.ID,
		Email:         user.TwitchAccount.Email,
		EmailVerified: user.TwitchAccount.EmailVerified,
		TwitchAccount: userv1Resource.TwitchAccount{
			TwitchID:       user.TwitchAccount.ID,
			TwitchUsername: user.TwitchAccount.Username,
			TwitchPicture:  user.TwitchAccount.Picture,
		},
		MCAccount: userv1Resource.MCAccount{
			MCUsername: user.MCAccount.Username,
			MCUUID:     user.MCAccount.ID,
		},
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, updatedUser)
}
