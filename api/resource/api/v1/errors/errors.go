package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	InternalServerErrorMsg = "Internal server error. Please try again. Report if the problem presists."
)

type InternalError struct {
	Message string
	Level   zapcore.Level
	Fields  []zap.Field
}

type APIError struct {
	StatusCode    int            `json:"code"`
	ErrorMessage  string         `json:"message"`
	InternalError *InternalError `json:"-"`
}

func (e *APIError) Error() string {
	return e.ErrorMessage
}

func (e *APIError) HTTPCode() int {
	return e.StatusCode
}

func (e *APIError) Log() {
	if e == nil {
		return
	}
	switch e.InternalError.Level {
	case zapcore.WarnLevel:
		zap.L().Warn(e.InternalError.Message, e.InternalError.Fields...)
	default:
		zap.L().Error(e.InternalError.Message, e.InternalError.Fields...)
	}
}

func NewInternalServerError(internalMsg string, err error) *gin.Error {
	return &gin.Error{
		Meta: APIError{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: InternalServerErrorMsg,
			InternalError: &InternalError{
				Message: internalMsg,
				Fields:  []zapcore.Field{zap.Error(err)},
			},
		},
	}
}
