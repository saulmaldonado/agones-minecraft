package v1Controllers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

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

func CreateJava(c *gin.Context) {
	v := c.GetString(jwt.ContextKey)
	userId := uuid.MustParse(v)

	var body gamev1Resource.CreateGameBody
	if err := c.ShouldBindJSON(&body); err != nil && err != io.EOF {
		c.Errors = append(c.Errors, errors.NewBadRequestError(err))
		return
	}

	gs := agones.NewJavaServer()

	if body.CustomSubdomain != nil {
		if ok := agones.Client().HostnameAvailable(agones.GetDNSZone(), *body.CustomSubdomain); !ok {
			c.Errors = append(c.Errors, errors.NewBadRequestError(fmt.Errorf("custom subdomain %s not available", *body.CustomSubdomain)))
			return
		}
		agones.SetHostname(gs, agones.GetDNSZone(), *body.CustomSubdomain)
	}

	gameServer, err := agones.Client().Create(gs)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	game := models.Game{
		ID:              uuid.MustParse(string(gameServer.UID)),
		UserID:          userId,
		Name:            gameServer.Name,
		CustomSubdomain: body.CustomSubdomain,
		Edition:         models.JavaEdition,
	}

	if err := gamev1Service.CreateGame(&game); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		agones.Client().Delete(gameServer.Name)
		return
	}

	createdGame := gamev1Resource.Game{
		ID:        game.ID,
		UserID:    game.UserID,
		Name:      game.Name,
		DNSRecord: agones.GetHostname(gameServer),
		Edition:   game.Edition,
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

	gs := agones.NewJavaServer()

	if body.CustomSubdomain != nil {
		if ok := agones.Client().HostnameAvailable(agones.GetDNSZone(), *body.CustomSubdomain); !ok {
			c.Errors = append(c.Errors, errors.NewBadRequestError(fmt.Errorf("custom subdomain %s not available", *body.CustomSubdomain)))
			return
		}
		agones.SetHostname(gs, agones.GetDNSZone(), *body.CustomSubdomain)
	}

	gameServer, err := agones.Client().Create(gs)
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}

	game := models.Game{
		ID:              uuid.MustParse(string(gameServer.UID)),
		UserID:          userId,
		Name:            gameServer.Name,
		CustomSubdomain: body.CustomSubdomain,
		Edition:         models.BedrockEdition,
	}

	if err := gamev1Service.CreateGame(&game); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		agones.Client().Delete(gameServer.Name)
		return
	}

	createdGame := gamev1Resource.Game{
		ID:        game.ID,
		UserID:    game.UserID,
		Name:      game.Name,
		DNSRecord: agones.GetHostname(gameServer),
		Edition:   game.Edition,
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
