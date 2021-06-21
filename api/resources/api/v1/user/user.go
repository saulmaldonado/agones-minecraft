package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID      `json:"id"`
	Email         string         `json:"email"`
	EmailVerified bool           `json:"emailVerified"`
	TwitchAccount *TwitchAccount `json:"twitchAccount"`
	MCAccount     *MCAccount     `json:"mcAccount"`
	LastLogin     time.Time      `json:"lastLogin"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

type TwitchAccount struct {
	TwitchUsername string `json:"twitchUsername"`
	TwitchID       string `json:"twitchId"`
	TwitchPicture  string `json:"twitchPicture"`
}

type MCAccount struct {
	MCUsername string    `json:"mcUsername"`
	MCUUID     uuid.UUID `json:"mcUuid"`
}

type EditUserBody struct {
	MCUsername string `json:"mcUsername" binding:"required,mcusername"`
}
