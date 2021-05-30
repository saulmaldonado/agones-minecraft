package user

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"agones-minecraft/db"
	"agones-minecraft/models"
)

var (
	ErrUserRecordNotChanged error = errors.New("user record not changed")
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

		oldTokens <- *foundUser.TwitchToken.TwitchAccessToken
		oldTokens <- *foundUser.TwitchToken.TwitchRefreshToken

		if err := updateUserIfChanged(tx, user, &foundUser); err != nil {
			if errors.Is(err, ErrUserRecordNotChanged) {
				if err := tx.Model(&foundUser).Select("last_login").Omit("updated_at").Updates(user).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		err = tx.Model(&foundUser.TwitchToken).Updates(&user.TwitchToken).Error
		*user = foundUser
		return err
	})
}

func UpdateUser(user *models.User, oldTokens chan string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		foundUser := models.User{
			ID: user.ID,
		}
		if err := tx.Joins("TwitchToken").First(&foundUser).Error; err != nil {
			return err
		}

		if err := updateUserIfChanged(tx, user, &foundUser); err != nil &&
			!errors.Is(err, ErrUserRecordNotChanged) {
			return err
		}

		var err error
		if oldTokens != nil {
			oldTokens <- *user.TwitchToken.TwitchAccessToken
			oldTokens <- *user.TwitchToken.TwitchRefreshToken
			close(oldTokens)

			err = tx.Model(&foundUser.TwitchToken).Updates(&user.TwitchToken).Error
		}

		*user = foundUser
		return err
	})
}

func GetUserById(userId uuid.UUID, user *models.User) error {
	return db.DB().Joins("TwitchToken").First(user, userId).Error
}

func CreateUser(user *models.User) error {
	return db.DB().Create(user).Error
}

func EditUser(user *models.User) error {
	return db.DB().Model(user).Updates(user).First(user).Error
}

// Finds a users stored Twitch access and refresh tokens
func GetUserTwitchTokens(userId uuid.UUID, twitchToken *models.TwitchToken) error {
	return db.DB().Where("user_id = ?", userId).First(twitchToken).Error
}

func UpdateUserTwitchTokens(userId uuid.UUID, twitchToken *models.TwitchToken) error {
	if err := db.DB().Where("user_id = ?", userId).Updates(twitchToken).Error; err != nil {
		return err
	}
	return nil
}

func updateUserIfChanged(tx *gorm.DB, user *models.User, foundUser *models.User) error {
	if *user.Email != *foundUser.Email ||
		*user.TwitchUsername != *foundUser.TwitchUsername ||
		*user.TwitchPicture != *foundUser.TwitchPicture {
		return tx.Model(&foundUser).Omit(clause.Associations).Updates(user).Error
	}
	return ErrUserRecordNotChanged
}
