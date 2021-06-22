package user

import (
	"agones-minecraft/db"
	"agones-minecraft/models/v1/mc"
	"context"
	"time"

	userv1Model "agones-minecraft/models/v1/user"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

func CreateMCAccount(tx *pg.Tx, account *mc.MCAccount) error {
	_, err := tx.Model(account).Insert()
	return err
}

func UpsertUserMCAccount(user *userv1Model.User, userId uuid.UUID) error {
	return db.DB().RunInTransaction(context.Background(), func(t *pg.Tx) error {
		updatedAt := time.Now()

		_, err := t.Model(user.MCAccount).
			OnConflict("(user_id) WHERE deleted_at IS NULL DO UPDATE").
			Set("mc_id = EXCLUDED.mc_id, skin = EXCLUDED.skin, username = EXCLUDED.username, updated_at = ?", updatedAt).
			Insert()
		if err != nil {
			return err
		}

		updateUser := t.Model(user).
			Set("updated_at = ?", updatedAt).
			Where("id = ?", userId).
			Returning("*")

		return t.Model(user).
			Relation("TwitchAccount").
			Relation("MCAccount").
			WithUpdate("updateUser", updateUser).
			First()
	})
}
