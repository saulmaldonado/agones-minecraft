package user

import (
	"agones-minecraft/db"
	"agones-minecraft/models"

	"github.com/google/uuid"
)

func GetUserByEmail(email string, user *models.User) error {
	if res := db.DB().Where("email = ?", email).First(user); res.Error != nil {
		return res.Error
	}
	return nil
}

func GetUserByTwitchId(twitchId string, user *models.User) error {
	if res := db.DB().Where("twitch_id = ?", twitchId).First(user); res.Error != nil {
		return res.Error
	}
	return nil
}

func GetUserById(userId string, user *models.User) error {
	id, err := uuid.Parse(userId)
	if err != nil {
		return err
	}

	if res := db.DB().First(user, id); res.Error != nil {
		return res.Error
	}

	return nil
}

func CreateUser(user *models.User) error {
	if res := db.DB().Create(user); res.Error != nil {
		return res.Error
	}
	return nil
}

func EditUser(user *models.User) error {
	return db.DB().Model(user).Updates(user).First(user).Error
}
