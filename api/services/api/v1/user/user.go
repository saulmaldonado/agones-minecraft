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

func CreateUserWithTwitchAccount(t *pg.Tx, user *userv1Model.User) error {
	newUser := t.Model(user).
		Returning("*")

	newTwitchUser := t.Model(user.TwitchAccount).
		Value("user_id", "(SELECT id FROM u)").
		Returning("*")

	err := t.Model().
		WithInsert("u", newUser).
		WithInsert("twitch_account", newTwitchUser).
		Table("u").
		Select(user)

	return err
}

func UpsertUserByTwitchId(user *userv1Model.User, twitchId string) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		var foundUser userv1Model.User
		if err := GetUserByTwitchId(&foundUser, twitchId); err != nil {
			if err == pg.ErrNoRows {
				return CreateUserWithTwitchAccount(tx, user)
			}
			return err
		}

		if foundUser.TwitchAccount != nil {
			go RevokeOldTwitchTokens(*foundUser.TwitchAccount)
		}

		if user.TwitchAccount.HasChanged(foundUser.TwitchAccount) {
			if err := UpdateTwitchAccount(tx, user.TwitchAccount); err != nil {
				return err
			}
		} else {
			if err := UpdateTwitchAccountTokens(tx, user.TwitchAccount); err != nil {
				return err
			}
		}

		if err := UpdateLastLogin(tx, user, time.Now()); err != nil {
			return err
		}

		*user = foundUser
		return nil
	})
}

func GetUserByTwitchId(user *userv1Model.User, twitchId string) error {
	return db.DB().Model(user).
		Relation("TwitchAccount").
		Relation("MCAccount").
		Where("twitch_account.twitch_id = ?", twitchId).
		First()
}

func GetUserById(user *userv1Model.User, userId uuid.UUID) error {
	return db.DB().Model(user).
		Relation("TwitchAccount").
		Relation("MCAccount").
		Where("u.id = ?", userId).
		First()
}

func UpdateLastLogin(tx *pg.Tx, user *userv1Model.User, lastLogin time.Time) error {
	user.LastLogin = lastLogin
	_, err := tx.Model(user).WherePK().Set("last_login = ?last_login").Update()
	return err
}
