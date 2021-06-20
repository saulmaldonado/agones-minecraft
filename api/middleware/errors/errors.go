package errors

import (
	v1Err "agones-minecraft/errors/v1"
	apiErr "agones-minecraft/resources/api/v1/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		var errors []interface{}
		var statusCode int

		for _, err := range c.Errors {
			if e, ok := err.Meta.(*apiErr.APIError); ok {
				errors = append(errors, e)
				statusCode = e.HTTPCode()
			} else {
				e := apiErr.NewInternalServerError(err, v1Err.ErrUnknownID)
				errors = append(errors, e.Meta)
				statusCode = http.StatusInternalServerError
			}
		}

		if len(errors) > 0 {
			c.JSON(statusCode, gin.H{
				"errors": errors,
			})
		}

	}
}
