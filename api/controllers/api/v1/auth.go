package v1Controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"agones-minecraft/config"
	jwtmiddleware "agones-minecraft/middleware/jwt"
	"agones-minecraft/middleware/twitch"
	"agones-minecraft/models"
	"agones-minecraft/resource/api/v1/errors"
	userv1Service "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/auth/jwt"
	twitchauth "agones-minecraft/services/auth/twitch"
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

func Logout(c *gin.Context) {
	userId := c.GetString(twitch.SubjectKey)
	tokenId := c.GetString(jwtmiddleware.TokenIDKey)

	tokenStore := jwt.Get()

	if ok, _ := tokenStore.Exists(userId, tokenId); ok {
		if err := tokenStore.Delete(userId); err != nil {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
			return
		}
	}

	err := func() error {
		var twitchTokens models.TwitchToken
		if err := userv1Service.GetUserTwitchTokens(uuid.MustParse(userId), &twitchTokens); err != nil {
			return err
		}

		clientId, _, _ := config.GetTwichCreds()
		errs := twitchauth.RevokeTokens(*twitchTokens.TwitchAccessToken, *twitchTokens.TwitchRefreshToken, clientId)
		for _, e := range errs {
			zap.L().Warn("error invalidating old tokens", zap.Error(e))
		}
		return nil
	}()

	if err != nil && err != gorm.ErrRecordNotFound {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	c.Status(http.StatusNoContent)
}
