package jwt

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwt"

	"agones-minecraft/resource/api/v1/errors"
	jwtService "agones-minecraft/services/auth/jwt"
)

const (
	HeaderKey  string = "Authorization"
	ContextKey string = "JWT_SUBJECT"
	TokenIDKey string = "JWT_ID"
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
			c.Errors = append(c.Errors, errors.NewUnauthorizedError(err))
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
		return nil, fmt.Errorf("missing access token in Authorization header")
	}

	token, err := jwtService.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse token")
	}

	val, _ := token.Get(jwtService.RefreshKey)

	if val.(bool) {
		return nil, fmt.Errorf("token identified as refresh token")
	}

	if err := jwtService.ValidateToken(token); err != nil {
		return nil, fmt.Errorf("invalid access token")
	}

	if err := jwtService.VerifyAccessToken(tokenString); err != nil {
		return nil, fmt.Errorf("unable to verify access token")
	}
	return token, nil
}
