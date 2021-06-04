package twitch

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TwitchToken struct {
	gorm.Model
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID             uuid.UUID
	TwitchAccessToken  *string `gorm:"not null"`
	TwitchRefreshToken *string `gorm:"not null"`
}

func (u *TwitchToken) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		u.ID = id
	}
	return nil
}
