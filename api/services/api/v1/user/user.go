package user

import (
	"context"
	"errors"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"

	"agones-minecraft/db"
	userv1Model "agones-minecraft/models/v1/user"
)

var (
	ErrUserRecordNotChanged error = errors.New("user record not changed")
)

func GetUserByTwitchId(user *userv1Model.User, twitchId string) error {
	return db.DB().Model(user).
		Relation("TwitchAccount").
		Relation("MCAccount").
		Where("twitch_account.id = ?", twitchId).
		First()
}

func GetUserById(user *userv1Model.User, userId uuid.UUID) error {
	return db.DB().Model(user).
		Relation("TwitchAccount").
		Relation("MCAccount").
		Where("u.id = ?", userId).
		First()
}

func CreateUser(user *userv1Model.User) error {
	return db.DB().RunInTransaction(context.Background(), func(t *pg.Tx) error {
		if _, err := t.Model(user).Insert(); err != nil {
			return err
		}

		user.TwitchAccount.UserID = user.ID
		if _, err := t.Model(user.TwitchAccount).Insert(); err != nil {
			return err
		}

		return nil
	})
}

func UpsertUserByTwitchId(user *userv1Model.User, twitchId string) error {
	return db.DB().RunInTransaction(context.Background(), func(t *pg.Tx) error {
		var foundUser userv1Model.User
		if err := GetUserByTwitchId(&foundUser, twitchId); err != nil {
			if err == pg.ErrNoRows {
				return CreateUser(user)
			}
			return err
		}

		if foundUser.TwitchAccount != nil {
			go RevokeOldTwitchTokens(*foundUser.TwitchAccount)
		}

		if user.TwitchAccount.HasChanged(foundUser.TwitchAccount) {
			if err := UpdateTwitchAccount(user.TwitchAccount); err != nil {
				return err
			}
		}

		if err := updateLastLogin(t, user, time.Now()); err != nil {
			return err
		}

		*user = foundUser
		return nil
	})
}

func updateLastLogin(tx *pg.Tx, user *userv1Model.User, lastLogin time.Time) error {
	user.LastLogin = lastLogin
	_, err := tx.Model(user).WherePK().Set("last_login = ?last_login").Update()
	return err
}
