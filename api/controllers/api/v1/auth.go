package v1Controllers

import (
	"agones-minecraft/resource/api/v1/errors"
	"agones-minecraft/services/auth/jwt"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VerifyBody struct {
	RefreshToken string `json:"refreshToken"`
}

func Refresh(c *gin.Context) {
	var body VerifyBody
	if err := c.BindJSON(&body); err != nil {
		c.Errors = append(c.Errors, errors.NewBadRequestError(err))
		return
	}

	token, err := jwt.ParseToken(body.RefreshToken)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewBadRequestError(fmt.Errorf("unable to parse token")))
		return
	}

	v, _ := token.Get(jwt.RefreshKey)

	if !v.(bool) {
		c.Errors = append(c.Errors, errors.NewBadRequestError(fmt.Errorf("token identified as access token")))
		return
	}

	if err := jwt.ValidateToken(token); err != nil {
		c.Errors = append(c.Errors, errors.NewUnauthorizedError(err))
		return
	}

	if err := jwt.VerifyRefreshToken(body.RefreshToken); err != nil {
		c.Errors = append(c.Errors, errors.NewUnauthorizedError(fmt.Errorf("unable to verify refresh token")))
		return
	}

	userId := token.Subject()

	tokenStore := jwt.Get()
	ok, err := tokenStore.Exists(userId, token.JwtID())
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}
	if !ok {
		c.Errors = append(c.Errors, errors.NewUnauthorizedError(fmt.Errorf("invalidated refresh token")))
		return
	}

	tokens, err := jwt.NewTokens(userId)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	if err := tokenStore.Set(userId, tokens.TokenId, tokens.RefreshTokenExp); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
	})
}
