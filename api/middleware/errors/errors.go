package errors

import (
	"agones-minecraft/resource/api/v1/errors"

	"github.com/gin-gonic/gin"
)

func HandleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if err := c.Errors.Last(); err != nil {
			if e, ok := err.Err.(*errors.APIError); ok {
				c.JSON(e.StatusCode, e)
			}
		}
	}
}
