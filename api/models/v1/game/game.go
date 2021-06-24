package game

import (
	"fmt"

	"github.com/google/uuid"

	"agones-minecraft/models/v1/model"
)

type Edition string
type GameState string

const (
	JavaEdition    Edition = "java"
	BedrockEdition Edition = "bedrock"

	On  GameState = "ON"
	Off GameState = "OFF"
)

type Game struct {
	model.Model
	UserID  uuid.UUID `pg:"type:uuid,notnull,unique"`
	Name    string    `pg:"type:varchar(60)"`
	MOTD    string    `pg:"type:varchar(59)"`
	Slots   int       `pg:"default:10"`
	Address string    `pg:"type:varchar(63),notnull"`
	Edition Edition   `pg:"type:varchar(25),notnull"`
	State   GameState `pg:"type:varchar(25),default:off,notnull"`
}

func (g *Game) GetResourceName() string {
	return fmt.Sprintf("%s.%s", g.UserID.String(), g.Name)
}
