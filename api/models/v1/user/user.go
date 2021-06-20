package user

import (
	"time"

	"agones-minecraft/models/v1/game"
	"agones-minecraft/models/v1/mc"
	"agones-minecraft/models/v1/model"
	"agones-minecraft/models/v1/twitch"
)

type User struct {
	model.Model
	tableName     struct{}              `pg:"alias:u"`
	LastLogin     time.Time             `pg:"default:now();notnull"`
	TwitchAccount *twitch.TwitchAccount `pg:"rel:belongs-to"`
	MCAccount     *mc.MCAccount         `pg:"rel:belongs-to"`
	Games         []*game.Game          `pg:"rel:has-many"`
}
