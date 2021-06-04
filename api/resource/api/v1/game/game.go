package game

import (
	"agones-minecraft/models"
	"time"

	"github.com/google/uuid"
)

type CreateGameBody struct {
	CustomSubdomain *string `json:"customSubdomain" binding:"omitempty,hostname_rfc1123"`
}

type Game struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"userId"`
	Name      string           `json:"name"`
	DNSRecord string           `json:"dnsRecord"`
	Edition   models.Edition   `json:"edition"`
	State     models.GameState `json:"state"`
	CreatedAt time.Time        `json:"createdAt"`
}

type GameStatus struct {
	ID       uuid.UUID        `json:"id"`
	Name     string           `json:"name"`
	State    models.GameState `json:"state"`
	Edition  models.Edition   `json:"edition"`
	Hostname *string          `json:"hostname"`
	Address  *string          `json:"address"`
	Port     *int32           `json:"port"`
}
