package v1Controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"agones-minecraft/resource/api/v1/errors"
	"agones-minecraft/services/k8s/agones"
)

func ListGames(c *gin.Context) {
	gameServers, err := agones.Client().List()
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}
	c.JSON(http.StatusOK, gameServers.Items)
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
	gameServer, err := agones.Client().Create(agones.NewJavaServer())
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}
	c.JSON(http.StatusCreated, gameServer)
}

func CreateBedrock(c *gin.Context) {
	gameServer, err := agones.Client().Create(agones.NewBedrockServer())
	if err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}
	c.JSON(http.StatusCreated, gameServer)
}

func DeleteGame(c *gin.Context) {
	name := c.Param("name")
	if err := agones.Client().Delete(name); err != nil {
		c.Errors = append(c.Errors, errors.NewInternalServerError(err))
		return
	}
	c.Status(http.StatusNoContent)
}
