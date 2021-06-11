package v1Controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	v1Err "agones-minecraft/errors/v1"
	twitchv1Model "agones-minecraft/models/v1/twitch"
	userv1Model "agones-minecraft/models/v1/user"
	apiErr "agones-minecraft/resource/api/v1/errors"
	userv1Resource "agones-minecraft/resource/api/v1/user"
	userv1 "agones-minecraft/services/api/v1/user"
	"agones-minecraft/services/auth/jwt"
	sessionsauth "agones-minecraft/services/auth/sessions"
	"agones-minecraft/services/auth/twitch"
)

const (
	OIDCClaims string = `{ "id_token": { "email": null, "email_verified": null }, "userinfo": { "picture": null, "preferred_username": null } }`
)

var (
	ErrTwitchUnverifiedEmail error = errors.New("twitch email not verified")
)

func TwitchLogin(c *gin.Context) {
	config := twitch.NewTwitchConfig(twitch.TwitchOIDCProvider, oidc.ScopeOpenID, "user:read:email")

	state, err := sessionsauth.NewState()
	if err != nil {
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrNewState))
		return
	}

	if err := sessionsauth.AddStateFlash(c, state); err != nil {
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrEncodingCookie))
		return
	}

	claims := oauth2.SetAuthURLParam("claims", OIDCClaims)

	// Redirect to Twitch auth server
	http.Redirect(c.Writer, c.Request, config.AuthCodeURL(state, claims), http.StatusTemporaryRedirect)
}

func TwitchCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if ok, err := sessionsauth.VerifyStateFlash(c, state); err != nil {
		if ok {
			c.Errors = append(c.Errors, apiErr.NewInternalServerError(err, v1Err.ErrEncodingCookie))
		} else {
			switch err {
			case sessionsauth.ErrMissingState:
				c.Error(apiErr.NewUnauthorizedError(err, v1Err.ErrMissingState))
			case sessionsauth.ErrMissingStateChallenge:
				c.Error(apiErr.NewUnauthorizedError(err, v1Err.ErrMissingStateChallenge))
			case sessionsauth.ErrFailedStateChallenge:
				c.Error(apiErr.NewUnauthorizedError(err, v1Err.ErrFailedStateChallenge))
			}
		}
		return
	}

	token, err := twitch.NewToken(code)
	if err != nil {
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrTwitchTokenExchange))
		return
	}

	payload, err := twitch.GetPayload(token)
	if err != nil {
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrTwitchTokenPayload))
		return
	}

	if !payload.EmailVerified {
		c.Error(apiErr.NewBadRequestError(ErrTwitchUnverifiedEmail, v1Err.ErrTwitchUnverifiedEmail))
		return
	}

	user := userv1Model.User{
		Email:          &payload.Email,
		EmailVerified:  &payload.EmailVerified,
		TwitchID:       &payload.Sub,
		TwitchUsername: &payload.Username,
		TwitchPicture:  &payload.Picture,
		TwitchToken: twitchv1Model.TwitchToken{
			TwitchAccessToken:  &token.AccessToken,
			TwitchRefreshToken: &token.RefreshToken,
		},
		LastLogin: time.Now(),
	}

	if err := userv1.UpsertUserByTwitchId(&user, user.TwitchID); err != nil {
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrUpdatingUser))
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
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrGeneratingNewTokens))
		return
	}

	if err := jwt.Get().Set(foundUser.ID.String(), tokens.TokenId, tokens.RefreshTokenExp); err != nil {
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrSavingNewTokens))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
		"user":   foundUser,
	})
}
