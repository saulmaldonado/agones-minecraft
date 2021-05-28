package user

import (
	"agones-minecraft/db"
	"agones-minecraft/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetUserByEmail(email string, user *models.User) error {
	return db.DB().Where("email = ?", email).First(user).Error
}

func GetUserByTwitchId(user *models.User) error {
	return db.DB().Where("twitch_id = ?", user.TwitchID).Joins("TwitchToken").First(user).Error
}

func UpdateUserByTwitchId(user *models.User) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		tx = tx.Session(&gorm.Session{SkipHooks: true})
		err := tx.First(user, "twitch_id = ?", user.TwitchID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user.ID = uuid.New()
			user.TwitchToken.ID = uuid.New()
			tx.Create(user)
		} else {
			now := time.Now()
			user.UpdatedAt = now
			user.TwitchToken.UpdatedAt = now
			tx.Model(user).Select("updated_at", "email", "email_verified", "twitch_username").Updates(user)
			tx.Model(&user.TwitchToken).Where("user_id = ?", user.ID).Select("updated_at", "twitch_access_token", "twitch_refresh_token").Updates(&user.TwitchToken)
		}
		return nil
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
