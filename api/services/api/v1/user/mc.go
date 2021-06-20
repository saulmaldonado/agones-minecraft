package user

import (
	"agones-minecraft/db"
	"agones-minecraft/models/v1/mc"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

func CreateMCAccount(tx *pg.Tx, account *mc.MCAccount) error {
	_, err := tx.Model(account).Insert()
	return err
}

func EditMCAccount(account *mc.MCAccount, userId uuid.UUID) error {
	account.UpdatedAt = time.Now()
	_, err := db.DB().Model(account).Where("user_id = ?", userId).Update()
	return err
}
