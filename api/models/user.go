package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email              *string   `gorm:"not null;size:255"`
	EmailVerified      *bool     `gorm:"not null"`
	TwitchID           *string
	TwitchUsername     *string    `gorm:"size:25"`
	TwitchAccessToken  *string    `gorm:"not null"`
	TwitchRefreshToken *string    `gorm:"not null"`
	MCUsername         *string    `gorm:"size:16"`
	MCUUID             *uuid.UUID `gorm:"type:uuid"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	u.ID = id
	return nil
}
