package game

import (
	"time"

	"github.com/google/uuid"

	gamev1Model "agones-minecraft/models/v1/game"
)

type Status string

const (
	Online   Status = "online"
	Offline  Status = "offline"
	Starting Status = "starting"
	Stopping Status = "stopping"
)

type CreateGameBody struct {
	Address *string `json:"customSubdomain" binding:"omitempty,hostname_rfc1123"`
}

type Game struct {
	ID        uuid.UUID             `json:"id"`
	UserID    uuid.UUID             `json:"userId"`
	Name      string                `json:"name"`
	Address   string                `json:"address"`
	Edition   gamev1Model.Edition   `json:"edition"`
	State     gamev1Model.GameState `json:"state"`
	CreatedAt time.Time             `json:"createdAt"`
}

type GameStatus struct {
	ID       uuid.UUID           `json:"id"`
	UserID   uuid.UUID           `json:"-"`
	Name     string              `json:"name"`
	Status   Status              `json:"status"`
	Edition  gamev1Model.Edition `json:"edition"`
	Hostname *string             `json:"hostname"`
	Address  *string             `json:"address"`
	Port     *int32              `json:"port"`
}
