package model

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	ID        uuid.UUID `pg:"type:uuid,default:uuid_generate_v4()"`
	CreatedAt time.Time `pg:"default:now(),notnull"`
	UpdatedAt time.Time `pg:"default:now(),notnull"`
	DeletedAt time.Time `pg:",soft_delete"`
}
