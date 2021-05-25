package jwt

import (
	"agones-minecraft/config"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
)

const (
	// Key for the custom claims "refresh"
	RefreshKey string = "refresh"
	// "iss"
	Issuer string = "agones-minecraft"
	// "exp" timeout for access tokens
	AccessTokenTimeout time.Duration = time.Hour * 4
	// "exp" timeout for refresh tokens
	RefreshTokenTimeout time.Duration = time.Hour * 24
)

// Access and Refresh token pair
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Generates new refresh and access tokens for the given userId with the same iss, iat, id, and sub claims.
// exp for access token is 4h and for refresh token it is 24h
// sub is the given userId
// "refresh" is a custom claim to identify between refresh and access token
func NewTokens(userId string) (*TokenPair, error) {
	a := jwt.New()
	a.Set(jwt.IssuerKey, Issuer)
	a.Set(jwt.IssuedAtKey, time.Now().Unix())
	a.Set(jwt.JwtIDKey, uuid.NewString())
	a.Set(jwt.SubjectKey, userId)

	r, err := a.Clone()
	if err != nil {
		return nil, err
	}

	a.Set(RefreshKey, false)
	r.Set(RefreshKey, true)

	a.Set(jwt.ExpirationKey, time.Now().Add(AccessTokenTimeout))
	r.Set(jwt.ExpirationKey, time.Now().Add(RefreshTokenTimeout))

	accessTokenKey := config.GetJWTSecret()
	// TODO: have a different refresh token key
	refreshTokenKey := config.GetJWTSecret()

	accessToken, err := jwt.Sign(a, jwa.HS256, []byte(accessTokenKey))
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.Sign(r, jwa.HS256, []byte(refreshTokenKey))
	if err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: string(accessToken), RefreshToken: string(refreshToken)}, err
}

// Will parse token string into jwt.Token without verification or validation
func ParseToken(token string) (jwt.Token, error) {
	return jwt.ParseString(token)
}

// Will validate a tokens claims
func ValidateToken(token jwt.Token) error {
	return jwt.Validate(token)
}

// Will validate an access token using the access token key
func VerifyAccessToken(token string) error {
	_, err := jws.Verify([]byte(token), jwa.HS256, []byte(config.GetJWTSecret()))
	return err
}

// Will validate an access token using the refresh token key
func VerifyRefreshToken(token string) error {
	// TODO: change key for refresh tokens
	_, err := jws.Verify([]byte(token), jwa.HS256, []byte(config.GetJWTSecret()))
	return err
}
