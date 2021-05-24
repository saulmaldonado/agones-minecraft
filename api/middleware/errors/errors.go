package errors

import (
	v1Error "agones-minecraft/resource/api/v1/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func HandleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if err := c.Errors.Last(); err != nil {
			switch e := err.Meta.(type) {
			case *v1Error.APIError:
				c.JSON(e.StatusCode, e)
				e.Log()
			default:
				c.JSON(http.StatusInternalServerError, v1Error.APIError{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: err.Error(),
				})
				zap.L().Error("internal server error", zap.Error(err))
			}
		}
	}
}
