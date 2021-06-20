package twitch

import (
	"agones-minecraft/models/v1/model"

	"github.com/google/uuid"
)

type TwitchAccount struct {
	model.Model
	ID            string    `pg:",pk"`
	Email         string    `pg:",notnull"`
	EmailVerified bool      `pg:",notnull"`
	UserID        uuid.UUID `pg:"type:uuid"`
	AccessToken   string    `pg:",notnull"`
	RefreshToken  string    `pg:",notnull"`
	Picture       string
	Username      string `pg:"type:varchar(25),notnull"`
}

func (t *TwitchAccount) HasChanged(compared *TwitchAccount) bool {
	return !(t.Email == compared.Email &&
		t.EmailVerified == compared.EmailVerified &&
		t.Picture == compared.Picture)
}
