package user

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"agones-minecraft/db"
	"agones-minecraft/models"
)

func GetUserByEmail(email string, user *models.User) error {
	return db.DB().Where("email = ?", email).First(user).Error
}

func GetUserByTwitchId(user *models.User) error {
	return db.DB().Where("twitch_id = ?", user.TwitchID).Joins("TwitchToken").First(user).Error
}

func UpsertUserByTwitchId(user *models.User, oldTokens chan string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		defer close(oldTokens)
		var foundUser models.User
		err := tx.Joins("TwitchToken").First(&foundUser, "twitch_id = ?", user.TwitchID).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(user).Error
		} else if err != nil {
			return err
		}

		oldTokens <- *user.TwitchToken.TwitchAccessToken
		oldTokens <- *user.TwitchToken.TwitchRefreshToken

		if *user.Email != *foundUser.Email ||
			*user.TwitchUsername != *foundUser.TwitchUsername ||
			*user.TwitchPicture != *foundUser.TwitchPicture {
			if err := tx.Model(&foundUser).Omit(clause.Associations).Updates(user).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&foundUser).Select("last_login").Omit("updated_at").Updates(user).Error; err != nil {
				return err
			}
		}

		err = tx.Model(&foundUser.TwitchToken).Updates(&user.TwitchToken).Error
		*user = foundUser
		return err
	})
}

func GetUserById(userId string, user *models.User) error {
	id, err := uuid.Parse(userId)
	if err != nil {
		return err
	}
	return db.DB().Joins("TwitchToken").First(user, id).Error
}

func CreateUser(user *models.User) error {
	return db.DB().Create(user).Error
}

func EditUser(user *models.User) error {
	return db.DB().Model(user).Updates(user).First(user).Error
}
