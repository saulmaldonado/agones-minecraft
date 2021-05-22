package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	TwitchUsername *string   `json:"twitchUsername"`
	TwitchID       *string   `json:"twitchId"`
	MCUsername     *string   `json:"mcUsername"`
	MCUUID         *string   `json:"mcUuid"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
