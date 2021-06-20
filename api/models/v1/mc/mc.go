package mc

import (
	"agones-minecraft/models/v1/model"

	"github.com/google/uuid"
)

type MCAccount struct {
	model.Model
	UserID   uuid.UUID `pg:"type:uuid,notnull,unique"`
	Username string    `pg:"type:varchar(16),notnull"`
	Skin     string
}
