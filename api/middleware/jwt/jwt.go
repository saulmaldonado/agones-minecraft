package jwt

import (
	"agones-minecraft/resource/api/v1/errors"
	"agones-minecraft/services/auth/jwt"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	HeaderKey  string = "Authorization"
	ContextKey string = "JWT_SUBJECT"
	TokenIDKey string = "JWT_ID"
)

// returns middleware that will parse JWT token in Authorization header, validate it, verify it
// and set the userId in the current context as "JWT_SUBJECT"
func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		v := c.GetHeader(HeaderKey)
		tokenString := strings.TrimSpace(strings.TrimPrefix(v, "Bearer"))

		if tokenString == "" {
			c.Errors = append(c.Errors, errors.NewUnauthorizedError(fmt.Errorf("missing access token in Authorization header")))
			c.Abort()
			return
		}

		token, err := jwt.ParseToken(tokenString)
		if err != nil {
			c.Errors = append(c.Errors, errors.NewUnauthorizedError(fmt.Errorf("unable to parse token")))
			c.Abort()
			return
		}

		val, _ := token.Get(jwt.RefreshKey)

		if val.(bool) {
			c.Errors = append(c.Errors, errors.NewUnauthorizedError(fmt.Errorf("token identified as refresh token")))
			c.Abort()
			return
		}

		if err := jwt.ValidateToken(token); err != nil {
			c.Errors = append(c.Errors, errors.NewUnauthorizedError(err))
			c.Abort()
			return
		}

		if err := jwt.VerifyAccessToken(tokenString); err != nil {
			c.Errors = append(c.Errors, errors.NewUnauthorizedError(fmt.Errorf("unable to verify access token")))
			c.Abort()
			return
		}

		// Set userId for request context
		c.Set(ContextKey, token.Subject())
		c.Set(TokenIDKey, token.JwtID())

		c.Next()
	}
}
