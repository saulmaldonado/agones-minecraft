package v1Controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	"agones-minecraft/services/auth/twitch"
)

func TwitchLogin(c *gin.Context) {
	config := twitch.NewTwitchConfig(twitch.TwitchOIDCProvider, oidc.ScopeOpenID, "user:read:email")

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

	config := twitch.NewTwitchConfig(twitch.TwitchOIDCProvider)

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	payload, err := twitch.GetPayload(token, config.ClientID)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	if !payload.EmailVerified {
		c.Errors = append(c.Errors, errors.NewBadRequestError(fmt.Errorf("twitch email not verified")))
		return
	}

	var statusCode int = http.StatusOK
	user := models.User{
		Email:          &payload.Email,
		EmailVerified:  &payload.EmailVerified,
		TwitchID:       &payload.Sub,
		TwitchUsername: &payload.Username,
		TwitchPicture:  &payload.Picture,
		TwitchToken: models.TwitchToken{
			TwitchAccessToken:  &token.AccessToken,
			TwitchRefreshToken: &token.RefreshToken,
		},
		LastLogin: time.Now(),
	}

	oldTokens := make(chan string, 2)
	go revokeOldTokens(config.ClientID, oldTokens)

	if err := userv1.UpsertUserByTwitchId(&user, oldTokens); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	foundUser := userv1Resource.User{
		ID:             user.ID,
		Email:          *user.Email,
		EmailVerified:  *user.EmailVerified,
		TwitchID:       user.TwitchID,
		TwitchUsername: user.TwitchUsername,
		TwitchPicture:  user.TwitchPicture,
		LastLogin:      user.LastLogin,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	tokens, err := jwt.NewTokens(foundUser.ID.String())
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	tokenStore := jwt.Get()
	if err := tokenStore.Set(foundUser.ID.String(), tokens.TokenId, tokens.RefreshTokenExp); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	c.JSON(statusCode, gin.H{
		"tokens": tokens,
		"user":   foundUser,
	})
}

// Helper func to revoke old tokens and log errors in a goroutine
func revokeOldTokens(clientId string, oldTokens chan string) {
	errs := twitch.RevokeTokens(<-oldTokens, <-oldTokens, clientId)
	for _, e := range errs {
		zap.L().Warn("error invalidating old tokens", zap.Error(e))
	}
}
