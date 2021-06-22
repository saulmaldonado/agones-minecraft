package mc

import (
	"agones-minecraft/models/v1/model"

	"github.com/google/uuid"
)

type MCAccount struct {
	model.Model
	MCID     uuid.UUID `pg:"mc_id,type:uuid,notnull,unique"`
	UserID   uuid.UUID `pg:"type:uuid,notnull,unique"`
	Username string    `pg:"type:varchar(16),notnull"`
	Skin     string
}
