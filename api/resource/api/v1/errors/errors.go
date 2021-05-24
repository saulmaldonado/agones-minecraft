package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	InternalServerErrorMsg = "Internal server error. Please try again. Report if the problem presists."
)

type APIError struct {
	StatusCode   int    `json:"code"`
	ErrorMessage string `json:"message"`
}

func (e *APIError) Error() string {
	return e.ErrorMessage
}

func (e *APIError) HTTPCode() int {
	return e.StatusCode
}

func NewInternalServerError(err error) *gin.Error {
	return &gin.Error{
		Err: err,
		Meta: APIError{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: InternalServerErrorMsg,
		},
	}
}
