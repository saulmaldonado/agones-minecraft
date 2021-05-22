package v1Controllers

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	sessionsauth "agones-minecraft/services/auth/sessions"
	twitchauth "agones-minecraft/services/auth/twitch"
)

func TwitchLogin(c *gin.Context) {
	config := twitchauth.NewTwitchConfig(twitchauth.TwitchOIDCProvider, oidc.ScopeOpenID, "user:read:email")

	state, err := sessionsauth.NewState()
	if err != nil {
		zap.L().Error("error generating new state", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := sessionsauth.AddStateFlash(c, state); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	claims := oauth2.SetAuthURLParam("claims", `{ "id_token": { "email": null, "email_verified": null }, "userinfo": { "picture": null, "preferred_username": null } }`)

	http.Redirect(c.Writer, c.Request, config.AuthCodeURL(state, claims), http.StatusTemporaryRedirect)
}

func TwitchCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	ok, err := sessionsauth.VerifyStateFlash(c, state)
	if err != nil {
		c.Status(http.StatusBadRequest)
		zap.L().Error("error verifying state", zap.Error(err))
		return
	}

	if !ok {
		c.Status(http.StatusUnauthorized)
		zap.L().Warn("failed state challenge")
		return
	}

	config := twitchauth.NewTwitchConfig(twitchauth.TwitchOIDCProvider)

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		zap.L().Error("error exchaning code for token", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)

	if !ok {
		zap.L().Error("missing id_token")
		c.Status(http.StatusInternalServerError)
		return
	}

	idToken, err := twitchauth.VerifyToken(config.ClientID, rawIDToken)
	if err != nil {
		zap.L().Error("error verifying token", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	var claims twitchauth.Claims

	if err := twitchauth.GetClaimsFromToken(idToken, &claims); err != nil {
		zap.L().Error("error extracting claims from token", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	var userInfo twitchauth.UserInfo

	if err := twitchauth.GetUserInfo(token.AccessToken, &userInfo); err != nil {
		zap.L().Error("error getting userinfo", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, token)
}
