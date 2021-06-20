package jwt

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwt"

	v1Err "agones-minecraft/errors/v1"
	apiErr "agones-minecraft/resources/api/v1/errors"
	jwtService "agones-minecraft/services/auth/jwt"
)

const (
	HeaderKey  string = "Authorization"
	ContextKey string = "JWT_SUBJECT"
	TokenIDKey string = "JWT_ID"
)

var (
	ErrMissingAccessToken        error = errors.New("missing access token in Authorization header")
	ErrAccessTokenParsing        error = errors.New("unable to parse token")
	ErrAccessTokenExpected       error = errors.New("token identified as access token expected access token")
	ErrInvalidAccessToken        error = errors.New("invalid access token")
	ErrUnableToVerifyAccessToken error = errors.New("unable to verify access token")
)

// returns middleware that will parse JWT token in Authorization header, validate it, verify it
// and set the userId in the current context as "JWT_SUBJECT".
// If authentication fails, the authorizer will return a 401 error
func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		v := c.GetHeader(HeaderKey)
		tokenString := strings.TrimSpace(strings.TrimPrefix(v, "Bearer"))

		token, err := authenticateWithToken(tokenString)
		if err != nil {
			switch err {
			case ErrMissingAccessToken:
				c.Error(apiErr.NewUnauthorizedError(ErrMissingAccessToken, v1Err.ErrMissingAccessToken))
			case ErrAccessTokenParsing:
				c.Error(apiErr.NewBadRequestError(ErrAccessTokenParsing, v1Err.ErrAccessTokenParsing))
			case ErrAccessTokenExpected:
				c.Error(apiErr.NewBadRequestError(ErrAccessTokenExpected, v1Err.ErrAccessTokenExpected))
			case ErrInvalidAccessToken:
				c.Error(apiErr.NewUnauthorizedError(ErrInvalidAccessToken, v1Err.ErrInvalidAccessToken))
			case ErrUnableToVerifyAccessToken:
				c.Error(apiErr.NewUnauthorizedError(ErrUnableToVerifyAccessToken, v1Err.ErrUnableToVerifyAccessToken))
			}
			c.Abort()
			return
		}

		// Set userId for request context
		c.Set(ContextKey, token.Subject())
		// Set tokenId for request context
		c.Set(TokenIDKey, token.JwtID())

		c.Next()
	}
}

// returns middleware that will parse JWT token in Authorization header, validate it, verify it
// and set the userId in the current context as "JWT_SUBJECT"
// If authentication fails, the authenticator will not set the userId for the context and will access the
// endpoint as a guest
func Authenticator() gin.HandlerFunc {
	return func(c *gin.Context) {
		v := c.GetHeader(HeaderKey)
		tokenString := strings.TrimSpace(strings.TrimPrefix(v, "Bearer"))

		token, err := authenticateWithToken(tokenString)
		if err == nil {
			// Set userId for request context
			c.Set(ContextKey, token.Subject())
			// Set tokenId for request context
			c.Set(TokenIDKey, token.JwtID())
		}

		c.Next()
	}
}

func authenticateWithToken(tokenString string) (jwt.Token, error) {
	if tokenString == "" {

		return nil, ErrMissingAccessToken
	}

	token, err := jwtService.ParseToken(tokenString)
	if err != nil {
		return nil, ErrAccessTokenParsing
	}

	val, _ := token.Get(jwtService.RefreshKey)

	if val.(bool) {
		return nil, ErrAccessTokenExpected
	}

	if err := jwtService.ValidateToken(token); err != nil {
		return nil, ErrInvalidAccessToken
	}

	if err := jwtService.VerifyAccessToken(tokenString); err != nil {
		return nil, ErrUnableToVerifyAccessToken
	}
	return token, nil
}
