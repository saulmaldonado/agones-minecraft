package v1Controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"agones-minecraft/models"
	"agones-minecraft/resource/api/v1/errors"
	userv1Resource "agones-minecraft/resource/api/v1/user"
	userv1 "agones-minecraft/services/api/v1/user"
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
		c.Errors = append(c.Errors,
			&gin.Error{
				Err: err,
				Meta: errors.APIError{
					StatusCode:   http.StatusBadRequest,
					ErrorMessage: "Error verifying authentication request. Make sure cookies are enabled.",
				},
			},
		)
		return
	}

	config := twitchauth.NewTwitchConfig(twitchauth.TwitchOIDCProvider)

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)

	if !ok {
		c.Errors = append(c.Errors, errors.NewInternalServerError(fmt.Errorf("id_token not included in token")))
		return
	}

	idToken, err := twitchauth.VerifyToken(config.ClientID, rawIDToken)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	var claims twitchauth.Claims

	if err := twitchauth.GetClaimsFromToken(idToken, &claims); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	var userInfo twitchauth.UserInfo

	if err := twitchauth.GetUserInfo(token.AccessToken, &userInfo); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	var user models.User
	var statusCode int = http.StatusOK

	if err := userv1.GetUserByEmail(claims.Email, &user); err != nil {
		user.Email = claims.Email
		user.TwitchUsername = userInfo.Username

		if err := userv1.CreateUser(&user); err != nil {
			c.Errors = append(c.Errors,
				&gin.Error{
					Err: err,
					Meta: errors.APIError{
						StatusCode:   http.StatusInternalServerError,
						ErrorMessage: errors.InternalServerErrorMsg,
					},
				},
			)
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

	c.JSON(statusCode, gin.H{
		"token": token,
		"user":  foundUser,
	})
}
