package v1Controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"agones-minecraft/resource/api/v1/errors"
	"agones-minecraft/services/auth/jwt"
)

type VerifyBody struct {
	RefreshToken string `json:"refreshToken"`
}

func Refresh(c *gin.Context) {
	var body VerifyBody
	if err := c.BindJSON(&body); err != nil {
		c.Errors = append(c.Errors, &gin.Error{
			Err:  err,
			Type: gin.ErrorTypeBind,
			Meta: errors.APIError{
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: err.Error(),
			},
		})
		return
	}

	jwtToken, err := jwt.GetRefreshToken(body.RefreshToken)
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{
			Err:  err,
			Type: gin.ErrorTypeBind,
			Meta: errors.APIError{
				StatusCode:   http.StatusUnauthorized,
				ErrorMessage: err.Error(),
			},
		})
		return
	}

	tokenClaims := jwtToken.Claims.(*jwt.Claims)
	accessToken, err := jwt.NewAccessToken(uuid.MustParse(tokenClaims.UserID))
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	refreshToken, err := jwt.NewRefreshToken(uuid.MustParse(tokenClaims.UserID))
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}
