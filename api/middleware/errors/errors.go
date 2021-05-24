package errors

import (
	v1Error "agones-minecraft/resource/api/v1/errors"

	"github.com/gin-gonic/gin"
)

func HandleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if err := c.Errors.Last(); err != nil {
			if e, ok := err.Meta.(v1Error.APIError); ok {
				c.JSON(e.StatusCode, e)
			}
		}
	}
}
