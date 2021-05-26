package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID  `json:"id"`
	Email          string     `json:"email"`
	EmailVerified  bool       `json:"emailVerified"`
	TwitchUsername *string    `json:"twitchUsername"`
	TwitchID       *string    `json:"twitchId"`
	MCUsername     *string    `json:"mcUsername"`
	MCUUID         *uuid.UUID `json:"mcUuid"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type EditUserBody struct {
	MCUsername string `json:"mcUsername" binding:"required,mcusername"`
}
