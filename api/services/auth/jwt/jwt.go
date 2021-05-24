package jwt

import (
	"agones-minecraft/config"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type claims struct {
	jwt.StandardClaims
	ID       string        `json:"id"`
	Provider loginProvider `json:"provider"`
}

type loginProvider string

const (
	twitch loginProvider = "twitch"
)

func NewAccessToken(id uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		ID:       id.String(),
		Provider: twitch,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 4).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	})
	return token.SignedString([]byte(config.GetJWTSecret()))
}

func NewRefreshToken(id uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		ID:       id.String(),
		Provider: twitch,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	})
	return token.SignedString([]byte(config.GetJWTSecret()))
}
