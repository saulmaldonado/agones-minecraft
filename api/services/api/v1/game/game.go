package game

import (
	"agones-minecraft/db"
	"context"
	"errors"

	v1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"

	gamev1Model "agones-minecraft/models/v1/game"
	gamev1Resource "agones-minecraft/resource/api/v1/game"
	"agones-minecraft/services/k8s/agones"
)

var (
	ErrSubdomainTaken     error = errors.New("custom subdomain not available")
	ErrGameServerNotFound error = errors.New("game server not found")
)

type ErrDeletingGameFromK8S struct {
	error
}

type ErrDeletingGameFromDB struct {
	error
}

func GetGameById(game *gamev1Model.Game, id uuid.UUID) error {
	game.ID = id
	return db.DB().Model(game).WherePK().First()
}

func GetGameByName(game *gamev1Model.Game, name string) error {
	return db.DB().Model(game).Where("name = ?", name).First()
}

func GetGameStatusByName(game *gamev1Resource.GameStatus, name string) error {
	var foundGame gamev1Model.Game
	if err := GetGameByName(&foundGame, name); err != nil {
		return err
	}

	*game = gamev1Resource.GameStatus{
		ID:      foundGame.ID,
		Name:    foundGame.Name,
		Status:  gamev1Resource.Offline,
		Edition: foundGame.Edition,
	}

	return nil
}

func GetGameByUserIdAndName(game *gamev1Model.Game, userId uuid.UUID, name string) error {
	return db.DB().Model(game).Where("name = ? AND user_id = ?", name, userId).First()
}

func CreateGame(game *gamev1Model.Game, gs *v1.GameServer) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		if ok := agones.Client().HostnameAvailable(agones.GetDNSZone(), game.Address); !ok {
			return ErrSubdomainTaken
		}
		agones.SetHostname(gs, agones.GetDNSZone(), game.Address)

		gameServer, err := agones.Client().CreateDryRun(gs)
		if err != nil {
			return err
		}

		// point to newly created gameserver obj
		*gs = *gameServer

		game.ID = uuid.MustParse(string(gs.UID))
		game.Name = gs.Name
		game.GameState = gamev1Model.On

		if _, err := db.DB().Model(game).Insert(); err != nil {
			return err
		}

		if _, err := agones.Client().Create(gs); err != nil {
			return err
		}

		return nil
	})
}

func DeleteGame(game *gamev1Model.Game, userId uuid.UUID, name string) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		if err := GetGameByUserIdAndName(game, userId, name); err != nil {
			if err == pg.ErrNoRows {
				return ErrGameServerNotFound
			}

			return &ErrDeletingGameFromDB{err}
		}

		if _, err := db.DB().Model(game).Delete(); err != nil {
			return &ErrDeletingGameFromDB{err}
		}

		if err := agones.Client().Delete(name); err != nil {
			return &ErrDeletingGameFromK8S{err}
		}

		return nil
	})
}

func UpdateGame(game *gamev1Model.Game) error {
	_, err := db.DB().Model(game).WherePK().Update()
	return err
}
