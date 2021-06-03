package game

import (
	"agones-minecraft/db"
	"agones-minecraft/models"

	"github.com/google/uuid"
)

func GetGameById(game *models.Game, ID uuid.UUID) error {
	game.ID = ID
	return db.DB().First(game).Error
}

func GetGameByUserIdAndName(game *models.Game, userId uuid.UUID, name string) error {
	return db.DB().Where("user_id = ? AND name = ?", userId, name).First(game).Error
}

func CreateGame(game *models.Game) error {
	return db.DB().Create(game).Error
}

func DeleteGame(game *models.Game) error {
	return db.DB().Delete(game).Error
}

func UpdateGame(game *models.Game) error {
	return db.DB().Model(game).Updates(game).First(game).Error
}
