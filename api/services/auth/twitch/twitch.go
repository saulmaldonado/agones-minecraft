package twitch

import (
	"agones-minecraft/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

const (
	oidcIssuer       = "https://id.twitch.tv/oauth2"
	userInfoEndpoint = "https://id.twitch.tv/oauth2/userinfo"
	validateEndpoint = "https://id.twitch.tv/oauth2/validate"
	revokeEndpoint   = "https://id.twitch.tv/oauth2/revoke"
)

type Payload struct {
	Claims
	UserInfo
}

type Claims struct {
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
	Aud           string `json:"aud"`
	Exp           int32  `json:"exp"`
	Iat           int32  `json:"iat"`
	Nonce         string `json:"nonce"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

type UserInfo struct {
	Picture  string `json:"picture"`
	Username string `json:"preferred_username"`
}

var ErrInvalidAccessToken error = errors.New("invalid twitch access token for user")
var ErrInvalidatedTokens error = errors.New("twitch credentials are invalid. login again to renew credentials and tokens")

var TwitchOIDCProvider *oidc.Provider

func NewTwitchConfig(provider *oidc.Provider, scopes ...string) *oauth2.Config {
	id, sec, re := config.GetTwichCreds()
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: sec,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  re,
		Scopes:       scopes,
	}
}

func Init() {
	TwitchOIDCProvider = NewODICProvider()
}

func NewODICProvider() *oidc.Provider {
	if TwitchOIDCProvider != nil {
		return TwitchOIDCProvider
	}
	prov, err := oidc.NewProvider(context.Background(), oidcIssuer)
	if err != nil {
		log.Fatal(err)
	}
	return prov
}

func NewOIDCVerifier(provider *oidc.Provider, clientId string) *oidc.IDTokenVerifier {
	return provider.Verifier(&oidc.Config{ClientID: clientId})
}

func GetUserInfo(token string, userInfo *UserInfo) error {
	req, err := http.NewRequest("GET", userInfoEndpoint, nil)

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		var httpError struct {
			Msg string `json:"message"`
		}
		if err := json.NewDecoder(res.Body).Decode(&httpError); err != nil {
			return err
		}
		return fmt.Errorf(httpError.Msg)
	}

	if err := json.NewDecoder(res.Body).Decode(userInfo); err != nil {
		return err
	}
	return nil
}

func VerifyToken(clientId, rawIDToken string) (idToken *oidc.IDToken, err error) {
	verifier := NewOIDCVerifier(TwitchOIDCProvider, clientId)

	idToken, err = verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		return nil, err
	}
	return idToken, err
}

func GetClaimsFromToken(idToken *oidc.IDToken, claims *Claims) error {
	if err := idToken.Claims(&claims); err != nil {
		return err
	}
	return nil
}

func GetPayload(token *oauth2.Token, clientId string) (*Payload, error) {
	rawIDToken, ok := token.Extra("id_token").(string)

	if !ok {
		return nil, fmt.Errorf("id_token not included in token")
	}

	idToken, err := VerifyToken(clientId, rawIDToken)
	if err != nil {
		return nil, err
	}

	var claims Claims

	if err := GetClaimsFromToken(idToken, &claims); err != nil {
		return nil, err
	}

	var userInfo UserInfo

	if err := GetUserInfo(token.AccessToken, &userInfo); err != nil {
		return nil, err
	}

	return &Payload{
		Claims:   claims,
		UserInfo: userInfo,
	}, nil
}

func ValidateToken(accessToken string) error {
	req, err := http.NewRequest("GET", validateEndpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "OAuth "+accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		var httpError struct {
			Msg string `json:"message"`
		}
		if err := json.NewDecoder(res.Body).Decode(&httpError); err != nil {
			return err
		}
		if httpError.Msg == "invalid access token" {
			return ErrInvalidAccessToken
		}
		return errors.New(httpError.Msg)
	}

	return nil
}

func Refresh(refreshToken, clientId, clientSecret string) (*oauth2.Token, error) {
	config := NewTwitchConfig(TwitchOIDCProvider)
	endpoint := config.Endpoint.TokenURL

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}

	que := req.URL.Query()
	que.Add("grant_type", "refresh_token")
	que.Add("refresh_token", refreshToken)
	que.Add("client_id", clientId)
	que.Add("client_secret", clientSecret)

	req.URL.RawQuery = que.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode == 401 || res.StatusCode == 400 {
		return nil, ErrInvalidatedTokens
	}

	var tokens oauth2.Token

	if err := json.NewDecoder(res.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
}

// Revokes old access and refresh tokens provided by Twitch
func RevokeTokens(accessToken, refreshToken, clientId string) []error {
	var errors []error

	if err := RevokeToken(accessToken, clientId); err != nil {
		errors = append(errors, err)
	}
	if err := RevokeToken(refreshToken, clientId); err != nil {
		errors = append(errors, err)
	}

	return errors
}

func RevokeToken(token, clientId string) error {
	if token == "" {
		return nil
	}

	tokenReq, err := http.NewRequest("POST", revokeEndpoint, nil)
	if err != nil {
		return err
	}

	aQ := tokenReq.URL.Query()
	aQ.Add("client_id", clientId)
	aQ.Add("token", token)

	tokenReq.URL.RawQuery = aQ.Encode()

	res, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == 400 {
		var body struct {
			Message string
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			return err
		}
		if body.Message != "Invalid token" {
			return fmt.Errorf("error invalidating access token. message: %s", body.Message)
		}
	}
	return nil

}
