package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	v1Err "agones-minecraft/errors/v1"
)

const (
	InternalServerErrorMsg = "internal server error. try again later. report if problem presists"
)

type APIError struct {
	Message     string        `json:"message"`
	StatusCode  int           `json:"statusCode"`
	ReferenceID v1Err.ErrorID `json:"referenceId"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) HTTPCode() int {
	return e.StatusCode
}

func NewInternalServerError(err error, id v1Err.ErrorID) *gin.Error {
	return &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePrivate,
		Meta: &APIError{
			Message:     InternalServerErrorMsg,
			StatusCode:  http.StatusInternalServerError,
			ReferenceID: id,
		},
	}
}

func NewBadRequestError(err error, id v1Err.ErrorID) *gin.Error {
	return &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePublic,
		Meta: &APIError{
			Message:     err.Error(),
			StatusCode:  http.StatusBadRequest,
			ReferenceID: id,
		},
	}
}

func NewUnauthorizedError(err error, id v1Err.ErrorID) *gin.Error {
	return &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePublic,
		Meta: &APIError{
			Message:     err.Error(),
			StatusCode:  http.StatusUnauthorized,
			ReferenceID: id,
		},
	}
}

func NewNotFoundError(err error, id v1Err.ErrorID) *gin.Error {
	return &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePublic,
		Meta: &APIError{
			Message:     err.Error(),
			StatusCode:  http.StatusNotFound,
			ReferenceID: id,
		},
	}
}

func NewGoneError(err error, id v1Err.ErrorID) *gin.Error {
	return &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePublic,
		Meta: &APIError{
			Message:     err.Error(),
			StatusCode:  http.StatusGone,
			ReferenceID: id,
		},
	}
}

func NewValidationError(verrs validator.ValidationErrors, id v1Err.ErrorID) []*gin.Error {
	var errs []*gin.Error

	for _, verr := range verrs {
		errMsg := verr.ActualTag()
		if verr.Param() != "" {
			errMsg = fmt.Sprintf("%s=%s", errMsg, verr.Param())
		}

		err := errors.New(errMsg)

		errs = append(errs,
			&gin.Error{
				Err:  err,
				Type: gin.ErrorTypeBind,
				Meta: &APIError{Message: err.Error(), StatusCode: http.StatusBadRequest, ReferenceID: id}},
		)
	}
	return errs
}
