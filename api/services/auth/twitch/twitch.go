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

var ErrInvalidAccessToken error = errors.New("account invalidated with Twitch")

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

func NewODICProvider() *oidc.Provider {
	if TwitchOIDCProvider != nil {
		return TwitchOIDCProvider
	}
	prov, err := oidc.NewProvider(context.Background(), oidcIssuer)
	if err != nil {
		log.Fatal(err)
	}
	TwitchOIDCProvider = prov
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
