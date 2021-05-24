package jwt

import (
	"agones-minecraft/config"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.StandardClaims
	UserID   string        `json:"userId"`
	Provider loginProvider `json:"provider"`
	Refresh  bool          `json:"refresh"`
}

type loginProvider string

const (
	twitch loginProvider = "twitch"
)

func NewAccessToken(id uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:   id.String(),
		Provider: twitch,
		Refresh:  false,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 4).Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
		},
	})
	return token.SignedString([]byte(config.GetJWTSecret()))
}

func NewRefreshToken(id uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:   id.String(),
		Provider: twitch,
		Refresh:  true,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
		},
	})
	return token.SignedString([]byte(config.GetJWTSecret()))
}

func GetAccessToken(token string) (*jwt.Token, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if c, ok := token.Claims.(*Claims); ok && !c.Refresh {
			return []byte(config.GetJWTSecret()), nil
		}
		return nil, fmt.Errorf("token identified as refresh token")
	})
	if err != nil {
		return nil, err
	}
	return jwtToken, nil
}

func GetRefreshToken(token string) (*jwt.Token, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if c, ok := token.Claims.(*Claims); ok && c.Refresh {
			return []byte(config.GetJWTSecret()), nil
		}
		return nil, fmt.Errorf("token identified as access token")
	})
	if err != nil {
		return nil, err
	}
	return jwtToken, nil
}
