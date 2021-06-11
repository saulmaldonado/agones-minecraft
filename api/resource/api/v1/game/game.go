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
	CustomSubdomain *string `json:"customSubdomain" binding:"omitempty,hostname_rfc1123"`
}

type Game struct {
	ID        uuid.UUID           `json:"id"`
	UserID    uuid.UUID           `json:"userId"`
	Name      string              `json:"name"`
	DNSRecord string              `json:"dnsRecord"`
	Edition   gamev1Model.Edition `json:"edition"`
	Status    Status              `json:"status"`
	CreatedAt time.Time           `json:"createdAt"`
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
