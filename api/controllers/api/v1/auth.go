package v1Controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agones-minecraft/services/auth/sessions"
)

var (
	ErrMissingRefreshToken        error = errors.New("missing refresh token in Authorization header")
	ErrRefreshTokenParsing        error = errors.New("unable to parse token")
	ErrRefreshTokenExpected       error = errors.New("token identified as access token expected refresh token")
	ErrInvalidRefreshToken        error = errors.New("invalid refresh token")
	ErrUnableToVerifyRefreshToken error = errors.New("unable to verify refresh token")
)

// func Refresh(c *gin.Context) {
// 	header := c.GetHeader(jwtmiddleware.HeaderKey)
// 	tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))

// 	if tokenString == "" {
// 		c.Error(apiErr.NewUnauthorizedError(ErrMissingRefreshToken, v1Err.ErrMissingRefreshToken))
// 		return
// 	}

// 	token, err := jwt.ParseToken(tokenString)
// 	if err != nil {
// 		c.Error(apiErr.NewBadRequestError(ErrRefreshTokenParsing, v1Err.ErrRefreshTokenParsing))
// 		return
// 	}

// 	v, ok := token.Get(jwt.RefreshKey)
// 	if !ok {
// 		c.Error(apiErr.NewBadRequestError(ErrRefreshTokenParsing, v1Err.ErrRefreshTokenParsing))
// 		return
// 	}

// 	if !v.(bool) {
// 		c.Error(apiErr.NewBadRequestError(ErrRefreshTokenExpected, v1Err.ErrRefreshTokenExpected))
// 		return
// 	}

// 	if err := jwt.ValidateToken(token); err != nil {
// 		c.Error(apiErr.NewUnauthorizedError(err, v1Err.ErrInvalidRefreshToken))
// 		return
// 	}

// 	if err := jwt.VerifyRefreshToken(tokenString); err != nil {
// 		c.Error(apiErr.NewUnauthorizedError(ErrUnableToVerifyRefreshToken, v1Err.ErrUnableToVerifyRefreshToken))
// 		return
// 	}

// 	userId := token.Subject()

// 	tokenStore := jwt.Get()
// 	ok, err = tokenStore.Exists(userId, token.JwtID())
// 	if err != nil {
// 		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingTokens))
// 		return
// 	}
// 	if !ok {
// 		c.Error(apiErr.NewUnauthorizedError(err, v1Err.ErrInvalidRefreshToken))
// 		return
// 	}

// 	tokens, err := jwt.NewTokens(userId)
// 	if err != nil {
// 		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrGeneratingNewTokens))
// 		return
// 	}

// 	if err := tokenStore.Set(userId, tokens.TokenId, tokens.RefreshTokenExp); err != nil {
// 		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrSavingNewTokens))
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"tokens": tokens,
// 	})
// }

func Logout(c *gin.Context) {
	if err := sessions.DestroySession(c); err != nil {
		zap.L().Warn("error revoking twitch tokens", zap.Error(err))
	}

	c.Status(http.StatusNoContent)
}
