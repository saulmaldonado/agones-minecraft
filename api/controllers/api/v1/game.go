package v1Controllers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"agones-minecraft/middleware/jwt"
	"agones-minecraft/models"
	"agones-minecraft/resource/api/v1/errors"
	gamev1Resource "agones-minecraft/resource/api/v1/game"
	gamev1Service "agones-minecraft/services/api/v1/game"
	"agones-minecraft/services/k8s/agones"
)

func ListGames(c *gin.Context) {
	gameServers, err := agones.Client().List()
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}
	c.JSON(http.StatusOK, gameServers)
}

func GetGame(c *gin.Context) {
	name := c.Param("name")
	gameServer, err := agones.Client().Get(name)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewNotFoundError(fmt.Errorf("server %s not found", name)))
		return
	}
	c.JSON(http.StatusOK, gameServer)
}

func GetGameState(c *gin.Context) {
	name := c.Param("name")
	game := gamev1Resource.GameStatus{
		Name: name,
	}

	if gs, err := agones.Client().Get(name); err != nil {
		var foundGame models.Game
		if k8sErrors.IsNotFound(err) {
			if err := gamev1Service.GetGameByName(&foundGame, name); err != nil {
				if err == gorm.ErrRecordNotFound {
					c.Errors = append(c.Errors, errors.NewNotFoundError(fmt.Errorf("game server not found")))
					return
				} else {
					c.Errors = append(c.Errors, errors.NewInternalServerError(err))
				}
			}
			game.ID = foundGame.ID
			game.State = models.Offline
			game.Edition = foundGame.Edition
		} else {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
			return
		}
	} else {
		game.Edition = agones.GetEdition(gs)
		game.ID = uuid.MustParse(string(gs.UID))
		game.State = agones.GetState(gs)

		if v, ok := c.Get(jwt.ContextKey); ok {
			userId := v.(string)
			if userId == agones.GetUserId(gs) && !agones.IsBeforePodCreated(gs) {
				port := agones.GetPort(gs)

				host := agones.GetHostname(gs)
				game.Hostname = &host
				game.Address = &gs.Status.Address
				game.Port = &port
			}
		}
	}

	c.JSON(http.StatusOK, game)
}

func CreateJava(c *gin.Context) {
	v := c.GetString(jwt.ContextKey)
	userId := uuid.MustParse(v)

	var body gamev1Resource.CreateGameBody
	if err := c.ShouldBindJSON(&body); err != nil && err != io.EOF {
		c.Errors = append(c.Errors, errors.NewBadRequestError(err))
		return
	}

	game := models.Game{
		CustomSubdomain: body.CustomSubdomain,
		UserID:          userId,
		Edition:         models.JavaEdition,
	}
	gs := agones.NewJavaServer()
	agones.SetUserId(gs, userId) // Set userId label

	if err := gamev1Service.CreateGame(&game, gs); err != nil {
		if err == gamev1Service.ErrSubdomainTaken {
			c.Errors = append(c.Errors, errors.NewBadRequestError(err))
		} else {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		}
		return
	}

	createdGame := gamev1Resource.Game{
		ID:        game.ID,
		UserID:    game.UserID,
		Name:      game.Name,
		DNSRecord: agones.GetHostname(gs),
		Edition:   game.Edition,
		State:     models.Starting,
		CreatedAt: game.CreatedAt,
	}

	c.JSON(http.StatusCreated, createdGame)
}

func CreateBedrock(c *gin.Context) {
	v := c.GetString(jwt.ContextKey)
	userId := uuid.MustParse(v)

	var body gamev1Resource.CreateGameBody
	if err := c.ShouldBindJSON(&body); err != nil && err != io.EOF {
		c.Errors = append(c.Errors, errors.NewBadRequestError(err))
		return
	}

	game := models.Game{
		CustomSubdomain: body.CustomSubdomain,
		UserID:          userId,
		Edition:         models.BedrockEdition,
	}
	gs := agones.NewBedrockServer()
	agones.SetUserId(gs, userId) // Set userId label

	if err := gamev1Service.CreateGame(&game, gs); err != nil {
		if err == gamev1Service.ErrSubdomainTaken {
			c.Errors = append(c.Errors, errors.NewBadRequestError(err))
		} else {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		}
		return
	}

	createdGame := gamev1Resource.Game{
		ID:        game.ID,
		UserID:    game.UserID,
		Name:      game.Name,
		DNSRecord: agones.GetHostname(gs),
		Edition:   game.Edition,
		State:     models.Starting,
		CreatedAt: game.CreatedAt,
	}

	c.JSON(http.StatusCreated, createdGame)
}

func DeleteGame(c *gin.Context) {
	v := c.GetString(jwt.ContextKey)
	userId := uuid.MustParse(v)

	name := c.Param("name")

	var game models.Game

	if err := gamev1Service.GetGameByUserIdAndName(&game, userId, name); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Errors = append(c.Errors, errors.NewNotFoundError(fmt.Errorf("game server %s for user %s not found", name, userId)))
		} else {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		}
		return
	}

	if err := gamev1Service.DeleteGame(&game); err != nil {
		if err != nil {
			c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		}
		return
	}

	if err := agones.Client().Delete(name); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	c.Status(http.StatusNoContent)
}
