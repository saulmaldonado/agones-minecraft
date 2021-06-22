package v1Controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	v1Err "agones-minecraft/errors/v1"
	"agones-minecraft/middleware/session"
	mcv1Model "agones-minecraft/models/v1/mc"
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
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	var user userv1Model.User
	if err := userv1Service.GetUserById(&user, userId); err != nil {
		if err == pg.ErrNoRows {
			c.Error(apiErr.NewNotFoundError(ErrUserNotFound, v1Err.ErrUserNotFound))
			return
		}
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingUser))
		return
	}

	var mcAccount *userv1Resource.MCAccount

	if user.MCAccount != nil {
		mcAccount = &userv1Resource.MCAccount{
			MCUsername: user.MCAccount.Username,
			MCUUID:     user.MCAccount.ID,
		}
	}

	foundUser := userv1Resource.User{
		ID:            user.ID,
		Email:         user.TwitchAccount.Email,
		EmailVerified: user.TwitchAccount.EmailVerified,
		TwitchAccount: &userv1Resource.TwitchAccount{
			TwitchID:       user.TwitchAccount.TwitchID,
			TwitchUsername: user.TwitchAccount.Username,
			TwitchPicture:  user.TwitchAccount.Picture,
		},
		MCAccount: mcAccount,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, foundUser)
}

func EditMe(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

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
		switch e := err.(type) {
		case *mc.ErrMcUserNotFound:
			c.Error(apiErr.NewBadRequestError(e, v1Err.ErrMcUserNotFound))
		case *mc.ErrUnmarshalingMCAccountJSON:
			c.Error(apiErr.NewInternalServerError(e, v1Err.ErrUnmarshalingMCAccountJSON))
		default:
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingMcUser))
		}
		return
	}

	user := userv1Model.User{
		MCAccount: &mcv1Model.MCAccount{
			MCID:     mcUser.UUID,
			UserID:   userId,
			Username: mcUser.Username,
			Skin:     mcUser.Textures.Skin.URL,
		},
	}

	if err := userv1Service.UpsertUserMCAccount(&user, userId); err != nil {
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
		TwitchAccount: &userv1Resource.TwitchAccount{
			TwitchID:       user.TwitchAccount.TwitchID,
			TwitchUsername: user.TwitchAccount.Username,
			TwitchPicture:  user.TwitchAccount.Picture,
		},
		MCAccount: &userv1Resource.MCAccount{
			MCUsername: user.MCAccount.Username,
			MCUUID:     user.MCAccount.ID,
			Skin:       user.MCAccount.Skin,
		},
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, updatedUser)
}
