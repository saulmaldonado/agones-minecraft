package game

import (
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/google/uuid"

	gamev1Model "agones-minecraft/models/v1/game"
	"agones-minecraft/services/k8s/agones"
)

type Status string

const (
	Online   Status = "online"
	Offline  Status = "offline"
	Starting Status = "starting"
	Stopping Status = "stopping"
)

type CreateGameBody struct {
	Name      string `json:"name" binding:"required,min=3,max=25"`
	Subdomain string `json:"subdomain" binding:"required,hostname_rfc1123"`
}

type Game struct {
	ID        uuid.UUID                 `json:"id"`
	UserID    uuid.UUID                 `json:"-"`
	Name      string                    `json:"name"`
	Edition   gamev1Model.Edition       `json:"edition"`
	State     gamev1Model.GameState     `json:"state,omitempty"`
	Status    *agonesv1.GameServerState `json:"status"`
	Address   string                    `json:"address"`
	Port      *int32                    `json:"port,omitempty"`
	CreatedAt time.Time                 `json:"createdAt"`
}

// Merge fields of a non-nil game model and a non-nil Agones GameServer resource
// into a game api resource
func (game *Game) MergeGame(gameModel *gamev1Model.Game, gs *agonesv1.GameServer) {
	if gameModel != nil {
		game.ID = gameModel.ID
		game.Name = gameModel.Name
		game.UserID = gameModel.UserID
		game.Address = gameModel.Address
		game.Edition = gameModel.Edition
		game.State = gameModel.State
		game.CreatedAt = gameModel.CreatedAt
	}

	if gs != nil {
		game.Status = agones.GetStatus(gs)
		if game.Edition == gamev1Model.BedrockEdition {
			game.Port = agones.GetPort(gs)
		}
	}
}
