package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"agones-minecraft/db"
	userv1Model "agones-minecraft/models/v1/user"
)

var (
	ErrUserRecordNotChanged error = errors.New("user record not changed")
)

func GetUserByTwitchId(user *userv1Model.User) error {
	return db.DB().Where("twitch_id = ?", user.TwitchID).Joins("TwitchToken").First(user).Error
}

func GetUserById(userId uuid.UUID, user *userv1Model.User) error {
	return db.DB().Joins("TwitchToken").First(user, userId).Error
}

func CreateUser(user *userv1Model.User) error {
	return db.DB().Create(user).Error
}

func UpsertUserByTwitchId(user *userv1Model.User, twitchId *string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		var foundUser userv1Model.User
		err := tx.Joins("TwitchToken").First(&foundUser, "twitch_id = ?", twitchId).Error

		if err == gorm.ErrRecordNotFound {
			return tx.Create(user).Error
		} else if err != nil {
			return err
		}

		// Revoke old token in goroutine
		go RevokeOldTwitchTokens(&foundUser.TwitchToken)

		if err := updateUserIfChanged(tx, user, &foundUser); err != nil {
			if err == ErrUserRecordNotChanged {
				if err := updateLastLogin(tx, &foundUser, user.LastLogin); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		if err := UpdateUserTwitchTokens(foundUser.ID, &user.TwitchToken); err != nil {
			return err
		}

		*user = foundUser
		return nil
	})
}

func EditUser(user *userv1Model.User) error {
	return db.DB().Model(user).Updates(user).First(user).Error
}

func updateUserIfChanged(tx *gorm.DB, user *userv1Model.User, foundUser *userv1Model.User) error {
	if user.HasChanged(foundUser) {
		return tx.Model(&foundUser).Omit(clause.Associations).Updates(user).Error
	}
	return ErrUserRecordNotChanged
}

func updateLastLogin(tx *gorm.DB, user *userv1Model.User, lastLogin time.Time) error {
	return tx.Model(user).Omit("updated_at").Update("last_login", lastLogin).Error
}
