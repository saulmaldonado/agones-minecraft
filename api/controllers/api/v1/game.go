package v1Controllers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	v1Err "agones-minecraft/errors/v1"
	"agones-minecraft/middleware/session"
	gamev1Model "agones-minecraft/models/v1/game"
	apiErr "agones-minecraft/resources/api/v1/errors"
	gamev1Resource "agones-minecraft/resources/api/v1/game"
	gamev1Service "agones-minecraft/services/api/v1/game"
	"agones-minecraft/services/k8s/agones"
)

var (
	ErrGameNotFound error = errors.New("game server not found")
)

func ListGamesForUser(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	var games []*gamev1Model.Game

	if err := gamev1Service.ListGamesForUser(&games, userId); err != nil {
		c.Error(apiErr.NewInternalServerError(err, v1Err.ErrListingGames))
		return
	}

	gameServers := []*gamev1Resource.Game{}

	for _, game := range games {
		gameServers = append(gameServers, &gamev1Resource.Game{
			ID:        game.ID,
			Name:      game.Name,
			Address:   game.Address,
			Edition:   game.Edition,
			State:     game.State,
			CreatedAt: game.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gameServers)
}

func GetGame(c *gin.Context) {
	name := c.Param("name")
	gameServer, err := agones.Client().Get(name)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			c.Error(apiErr.NewNotFoundError(ErrGameNotFound, v1Err.ErrGameNotFound))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingGameServer))
		}
		return
	}
	c.JSON(http.StatusOK, gameServer)
}

func GetGameState(c *gin.Context) {
	name := c.Param("name")

	var game gamev1Resource.GameStatus

	if err := agones.GetGameStatusByName(&game, name); err != nil {
		if k8sErrors.IsNotFound(err) {
			if err := gamev1Service.GetGameStatusByName(&game, name); err != nil {
				if err == gorm.ErrRecordNotFound {
					c.Error(apiErr.NewNotFoundError(ErrGameNotFound, v1Err.ErrGameNotFound))
				} else {
					c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingGameServerFromDB))
				}
				return
			}
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrRetrievingGameServer))
			return
		}
	} else {
		v, ok := c.Get(session.SessionUserIDKey)
		if ok {
			userId := v.(uuid.UUID).String()
			if userId == game.UserID.String() {
				c.JSON(http.StatusOK, game)
				return
			}
		}
	}

	game.Address = nil
	game.Hostname = nil
	game.Port = nil

	c.JSON(http.StatusOK, game)
}

func CreateJava(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	var body gamev1Resource.CreateGameBody
	if err := c.ShouldBindJSON(&body); err != nil && err != io.EOF {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			c.Errors = append(c.Errors, apiErr.NewValidationError(verrs, v1Err.ErrCreateGameServerValidation)...)
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrMalformedJSON))
		}
		return
	}

	game := gamev1Model.Game{
		Address: *body.Address,
		UserID:  userId,
		Edition: gamev1Model.JavaEdition,
	}
	gs := agones.NewJavaServer()
	agones.SetUserId(gs, userId) // Set userId label

	if err := gamev1Service.CreateGame(&game, gs); err != nil {
		if err == gamev1Service.ErrSubdomainTaken {
			c.Error(apiErr.NewBadRequestError(err, v1Err.ErrSubdomainTaken))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrCreatingGame))
		}
		return
	}

	createdGame := gamev1Resource.Game{
		ID:        game.ID,
		UserID:    game.UserID,
		Name:      game.Name,
		Address:   agones.GetHostname(gs),
		Edition:   game.Edition,
		State:     game.State,
		CreatedAt: game.CreatedAt,
	}

	c.JSON(http.StatusCreated, createdGame)
}

func CreateBedrock(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	var body gamev1Resource.CreateGameBody
	if err := c.ShouldBindJSON(&body); err != nil && err != io.EOF {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			c.Errors = append(c.Errors, apiErr.NewValidationError(verrs, v1Err.ErrCreateGameServerValidation)...)
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrMalformedJSON))
		}
		return
	}

	game := gamev1Model.Game{
		Address: *body.Address,
		UserID:  userId,
		Edition: gamev1Model.BedrockEdition,
	}
	gs := agones.NewBedrockServer()
	agones.SetUserId(gs, userId) // Set userId label

	if err := gamev1Service.CreateGame(&game, gs); err != nil {
		if err == gamev1Service.ErrSubdomainTaken {
			c.Error(apiErr.NewBadRequestError(err, v1Err.ErrSubdomainTaken))
		} else {
			c.Error(apiErr.NewInternalServerError(err, v1Err.ErrCreatingGame))
		}
		return
	}
	createdGame := gamev1Resource.Game{
		ID:        game.ID,
		UserID:    game.UserID,
		Name:      game.Name,
		Address:   agones.GetHostname(gs),
		Edition:   game.Edition,
		State:     game.State,
		CreatedAt: game.CreatedAt,
	}

	c.JSON(http.StatusCreated, createdGame)
}

func DeleteGame(c *gin.Context) {
	v, _ := c.Get(session.SessionUserIDKey)
	userId := v.(uuid.UUID)

	name := c.Param("name")

	var game gamev1Model.Game

	if err := gamev1Service.DeleteGame(&game, userId, name); err != nil {
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
