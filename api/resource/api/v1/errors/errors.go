package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	InternalServerErrorMsg = "Internal server error. Please try again. Report if the problem presists."
)

type APIError struct {
	Err          error  `json:"-"`
	ErrorMessage string `json:"errorMessage"`
	StatusCode   int    `json:"statusCode"`
}

func (e *APIError) Error() string {
	return e.ErrorMessage
}

func (e *APIError) HTTPCode() int {
	return e.StatusCode
}

func NewInternalServerError(err error) *gin.Error {
	return &gin.Error{
		Err: &APIError{
			Err:          err,
			ErrorMessage: InternalServerErrorMsg,
			StatusCode:   http.StatusInternalServerError,
		},
		Type: gin.ErrorTypeAny,
	}
}

func NewBadRequestError(err error) *gin.Error {
	return &gin.Error{
		Err: &APIError{
			Err:          nil,
			ErrorMessage: err.Error(),
			StatusCode:   http.StatusBadRequest,
		},
		Type: gin.ErrorTypeAny,
	}
}

func NewUnauthorizedError(err error) *gin.Error {
	return &gin.Error{
		Err: &APIError{
			Err:          nil,
			ErrorMessage: err.Error(),
			StatusCode:   http.StatusUnauthorized,
		},
		Type: gin.ErrorTypeAny,
	}
}

func NewNotFoundError(err error) *gin.Error {
	return &gin.Error{
		Err: &APIError{
			Err:          nil,
			ErrorMessage: err.Error(),
			StatusCode:   http.StatusNotFound,
		},
		Type: gin.ErrorTypeAny,
	}
}
