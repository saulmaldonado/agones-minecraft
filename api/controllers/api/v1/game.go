package v1Controllers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	v1Err "agones-minecraft/errors/v1"
	"agones-minecraft/middleware/session"
	gamev1Model "agones-minecraft/models/v1/game"
	apiErr "agones-minecraft/resources/api/v1/errors"
	gamev1Resource "agones-minecraft/resources/api/v1/game"
	gamev1Service "agones-minecraft/services/api/v1/game"
)

var (
	ErrGameNotFound       error = errors.New("game server not found")
	ErrMissingRequestBody error = errors.New("missing request body")
)

func ListGamesForUser(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	games := []*gamev1Resource.Game{}

	if err := gamev1Service.ListGamesForUser(&games, userId); err != nil {
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrListingGames))
		return
	}

	c.JSON(http.StatusOK, games)
}

func GetGame(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	name := c.Param("name")

	var game gamev1Resource.Game

	if err := gamev1Service.GetGameByNameAndUserId(&game, name, userId); err != nil {
		if err == gamev1Service.ErrGameServerNotFound {
			c.Error(apiErr.NewNotFoundError(ErrGameNotFound, v1Err.ErrGameNotFound))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingGameServer))
		}
		return
	}

	c.JSON(http.StatusOK, game)
}

func CreateJava(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	var body gamev1Resource.CreateGameBody
	if err := c.ShouldBindJSON(&body); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			c.Errors = append(c.Errors, apiErr.NewValidationError(verrs, v1Err.ErrCreateGameServerValidation)...)
		} else if err == io.EOF {
			c.Error(apiErr.NewBadRequestError(ErrMissingRequestBody, v1Err.ErrMissingRequestBody))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrMalformedJSON))
		}
		return
	}

	var game gamev1Resource.Game

	if err := gamev1Service.CreateGame(&game, gamev1Model.JavaEdition, body, userId); err != nil {
		if err == gamev1Service.ErrSubdomainTaken {
			c.Error(apiErr.NewBadRequestError(err, v1Err.ErrSubdomainTaken))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrCreatingGame))
		}
		return
	}

	c.JSON(http.StatusCreated, game)
}

func CreateBedrock(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	var body gamev1Resource.CreateGameBody
	if err := c.ShouldBindJSON(&body); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			c.Errors = append(c.Errors, apiErr.NewValidationError(verrs, v1Err.ErrCreateGameServerValidation)...)
		} else if err == io.EOF {
			c.Error(apiErr.NewBadRequestError(ErrMissingRequestBody, v1Err.ErrMissingRequestBody))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrMalformedJSON))
		}
		return
	}

	var game gamev1Resource.Game

	if err := gamev1Service.CreateGame(&game, gamev1Model.BedrockEdition, body, userId); err != nil {
		if err == gamev1Service.ErrSubdomainTaken {
			c.Error(apiErr.NewBadRequestError(err, v1Err.ErrSubdomainTaken))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrCreatingGame))
		}
		return
	}

	c.JSON(http.StatusCreated, game)
}

func DeleteGame(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	name := c.Param("name")

	if err := gamev1Service.DeleteGame(userId, name); err != nil {
		if err == gamev1Service.ErrGameServerNotFound {
			c.Error(apiErr.NewNotFoundError(err, v1Err.ErrGameNotFound))
		} else {
			switch err.(type) {
			case *gamev1Service.ErrDeletingGameFromDB:
				c.Error(apiErr.NewInternalServerError(err, v1Err.ErrDeletingGameFromDB))
			case *gamev1Service.ErrDeletingGameFromK8S:
				c.Error(apiErr.NewInternalServerError(err, v1Err.ErrDeletingGameFromK8s))
			}
		}
		return
	}

	c.Status(http.StatusNoContent)
}
