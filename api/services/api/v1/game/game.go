package game

import (
	"agones-minecraft/db"
	"errors"

	v1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/google/uuid"

	gamev1Model "agones-minecraft/models/v1/game"
	"agones-minecraft/services/k8s/agones"
)

var (
	ErrSubdomainTaken error = errors.New("custom subdomain not available")
)

func GetGameById(game *gamev1Model.Game, ID uuid.UUID) error {
	game.ID = ID
	return db.DB().First(game).Error
}

func GetGameByName(game *gamev1Model.Game, name string) error {
	return db.DB().Where("name = ?", name).First(game).Error
}

func GetGameByUserIdAndName(game *gamev1Model.Game, userId uuid.UUID, name string) error {
	return db.DB().Where("user_id = ? AND name = ?", userId, name).First(game).Error
}

func CreateGame(game *gamev1Model.Game, gs *v1.GameServer) error {
	if game.CustomSubdomain != nil {
		if ok := agones.Client().HostnameAvailable(agones.GetDNSZone(), *game.CustomSubdomain); !ok {
			return ErrSubdomainTaken
		}
		agones.SetHostname(gs, agones.GetDNSZone(), *game.CustomSubdomain)
	}

	gameServer, err := agones.Client().Create(gs)
	if err != nil {
		return err
	}
	// point to newly created gameserver obj
	*gs = *gameServer

	game.ID = uuid.MustParse(string(gameServer.UID))
	game.Name = gameServer.Name
	game.GameState = gamev1Model.On

	if err := db.DB().Create(game).Error; err != nil {
		// attempt to revert server
		agones.Client().Delete(gameServer.Name)
		return err
	}

	return nil
}

func DeleteGame(game *gamev1Model.Game) error {
	return db.DB().Delete(game).Error
}

func UpdateGame(game *gamev1Model.Game) error {
	return db.DB().Model(game).Updates(game).First(game).Error
}
