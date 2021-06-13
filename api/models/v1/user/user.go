package user

import (
	"reflect"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"agones-minecraft/models/v1/game"
	"agones-minecraft/models/v1/twitch"
)

type User struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email          *string   `gorm:"not null;size:255"`
	EmailVerified  *bool     `gorm:"not null"`
	TwitchID       *string   `gorm:"uniqueIndex"`
	TwitchUsername *string   `gorm:"size:25"`
	TwitchPicture  *string
	TwitchToken    twitch.TwitchToken `gorm:"constraint:OnDelete:CASCADE"`
	MCUsername     *string            `gorm:"size:16"`
	MCUUID         *uuid.UUID         `gorm:"type:uuid"`
	Game           []game.Game
	LastLogin      time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	u.ID = id
	return nil
}

func (u *User) HasChanged(c *User) bool {
	return !(reflect.DeepEqual(u.Email, c.Email) &&
		reflect.DeepEqual(u.EmailVerified, c.EmailVerified) &&
		reflect.DeepEqual(u.TwitchUsername, c.TwitchUsername) &&
		reflect.DeepEqual(u.TwitchPicture, c.TwitchPicture))
}
