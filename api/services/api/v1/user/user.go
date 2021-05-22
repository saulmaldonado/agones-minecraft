package user

import (
	"agones-minecraft/db"
	"agones-minecraft/models"
)

func GetUserByEmail(email string, user *models.User) error {
	if res := db.DB().Where("email = ?", email).First(user); res.Error != nil {
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
