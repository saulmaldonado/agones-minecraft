package v1Controllers

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"agones-minecraft/models"
	"agones-minecraft/resource/api/v1/errors"
	userv1Resource "agones-minecraft/resource/api/v1/user"
	userv1 "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/auth/jwt"
	sessionsauth "agones-minecraft/services/auth/sessions"
	twitchauth "agones-minecraft/services/auth/twitch"
)

func TwitchLogin(c *gin.Context) {
	config := twitchauth.NewTwitchConfig(twitchauth.TwitchOIDCProvider, oidc.ScopeOpenID, "user:read:email")

	state, err := sessionsauth.NewState()
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	if err := sessionsauth.AddStateFlash(c, state); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}
	claims := oauth2.SetAuthURLParam("claims", `{ "id_token": { "email": null, "email_verified": null }, "userinfo": { "picture": null, "preferred_username": null } }`)
	http.Redirect(c.Writer, c.Request, config.AuthCodeURL(state, claims), http.StatusTemporaryRedirect)
}

func TwitchCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	ok, err := sessionsauth.VerifyStateFlash(c, state)
	if err != nil || !ok {
		c.Errors = append(c.Errors, errors.NewBadRequestError(err))
		return
	}

	config := twitchauth.NewTwitchConfig(twitchauth.TwitchOIDCProvider)

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	payload, err := twitchauth.GetPayload(token, config.ClientID)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	var user models.User
	var statusCode int = http.StatusOK

	if err := userv1.GetUserByEmail(payload.Email, &user); err != nil {
		user.Email = payload.Email
		user.TwitchUsername = payload.Username

		if err := userv1.CreateUser(&user); err != nil {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
			return
		}
		statusCode = http.StatusCreated
		zap.L().Info(
			"new user created",
			zap.String("id", user.ID.String()),
			zap.String("email", user.Email),
			zap.String("username", user.TwitchUsername),
		)
	}

	foundUser := userv1Resource.User{
		ID:             user.ID,
		Email:          user.Email,
		TwitchUsername: &user.TwitchUsername,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	accessToken, err := jwt.NewAccessToken(foundUser.ID)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	refreshToken, err := jwt.NewRefreshToken(foundUser.ID)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	c.JSON(statusCode, gin.H{
		"token": gin.H{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		},
		"user": foundUser,
	})
}
