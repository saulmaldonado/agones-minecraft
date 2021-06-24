package game

import (
	"agones-minecraft/db"
	"context"
	"errors"
	"fmt"
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	gamev1Model "agones-minecraft/models/v1/game"
	gamev1Resource "agones-minecraft/resources/api/v1/game"
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

func GetGameById(game *gamev1Resource.Game, id uuid.UUID) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		var foundGame gamev1Model.Game

		game.ID = id
		if err := tx.Model(game).WherePK().First(); err != nil {
			return err
		}

		gs, err := agones.Client().Get(foundGame.GetResourceName())
		if err == nil {
			return reconcileGameState(tx, &foundGame, gs)
		} else if !k8sErrors.IsNotFound(err) {
			return err
		}

		game.MergeGame(&foundGame, gs)

		return nil
	})
}

func GetGameByNameAndUserId(game *gamev1Resource.Game, name string, userId uuid.UUID) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		var foundGame gamev1Model.Game

		if err := getByNameAndUserId(tx, &foundGame, name, userId); err != nil {
			if err == pg.ErrNoRows {
				return ErrGameServerNotFound
			}
			return err
		}

		gs, err := agones.Client().GetForUser(foundGame.GetResourceName(), userId)
		if err != nil {
			if !k8sErrors.IsNotFound(err) {
				return err
			}
		}

		if err := reconcileGameState(tx, &foundGame, gs); err != nil {
			return err
		}

		game.MergeGame(&foundGame, gs)
		return nil
	})
}

func ListGamesForUser(games *[]*gamev1Resource.Game, userId uuid.UUID) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		foundGames := []*gamev1Model.Game{}
		err := tx.Model(&foundGames).Where("user_id = ?", userId).Select()
		if err != nil {
			if err != pg.ErrNoRows {
				return err
			}
		}

		gsList, err := agones.Client().ListGamesForUser(userId.String())
		if err != nil {
			return err
		}

		if err := reconcileGameListStates(tx, foundGames, gsList); err != nil {
			return err
		}

		listedGames := make(map[string]*agonesv1.GameServer)
		for _, gs := range gsList {
			listedGames[agones.GetUUID(gs).String()] = gs
		}

		for _, foundGame := range foundGames {
			game := gamev1Resource.Game{}
			fmt.Println(listedGames)
			listedGame := listedGames[foundGame.ID.String()]
			game.MergeGame(foundGame, listedGame)
			*games = append(*games, &game)
		}

		return nil
	})
}

func CreateGame(game *gamev1Resource.Game, edition gamev1Model.Edition, body gamev1Resource.CreateGameBody, userId uuid.UUID) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		if !agones.Client().AddressAvailable(body.Subdomain) {
			return ErrSubdomainTaken
		}

		var gs *agonesv1.GameServer

		switch edition {
		case gamev1Model.BedrockEdition:
			gs = agones.NewBedrockServer()
		default:
			gs = agones.NewJavaServer()
		}

		agones.SetName(gs, userId, body.Name) // Set cluster unique GameServer name

		gameModel := gamev1Model.Game{
			Name:    agones.GetName(gs),
			State:   gamev1Model.On,
			Address: agones.NewAddress(body.Subdomain),
			UserID:  userId,
			Edition: gamev1Model.JavaEdition,
		}

		if _, err := tx.Model(&gameModel).Insert(); err != nil {
			return err
		}

		agones.SetUserId(gs, userId)                                // Set userId label
		agones.SetHostname(gs, agones.GetDNSZone(), body.Subdomain) // Set externalDNS annoataions
		agones.SetUUID(gs, gameModel.ID)                            // Set record uuid label

		newGs, err := agones.Client().Create(gs)
		if err != nil {
			return err
		}

		game.MergeGame(&gameModel, newGs)

		return nil
	})
}

func DeleteGame(userId uuid.UUID, name string) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		var foundGame gamev1Model.Game
		if err := getByNameAndUserId(tx, &foundGame, name, userId); err != nil {
			if err == pg.ErrNoRows {
				return ErrGameServerNotFound
			}
			return &ErrDeletingGameFromDB{err}
		}

		if _, err := tx.Model(&foundGame).WherePK().Delete(); err != nil {
			return &ErrDeletingGameFromDB{err}
		}

		if err := agones.Client().Delete(foundGame.GetResourceName()); err != nil {
			return &ErrDeletingGameFromK8S{err}
		}

		return nil
	})
}

func UpdateGame(game *gamev1Model.Game) error {
	_, err := db.DB().Model(game).WherePK().Update()
	return err
}

// Reconcile game state in database to match the state in cluster
func reconcileGameState(tx *pg.Tx, game *gamev1Model.Game, gs *agonesv1.GameServer) error {
	realState := agones.GetState(gs)
	var err error
	if realState != game.State {
		_, err = tx.Model(game).
			Set("state = ?", realState).
			Set("updated_at = ?", time.Now()).
			WherePK().
			Update()
	}
	return err
}

// Bulk reconcile game states in database to match the game states in the cluster
func reconcileGameListStates(tx *pg.Tx, games []*gamev1Model.Game, gsList []*agonesv1.GameServer) error {
	updates := []*gamev1Model.Game{}
	now := time.Now()
	realStates := make(map[string]gamev1Model.GameState)

	for _, gs := range gsList {
		realStates[agones.GetUUID(gs).String()] = agones.GetState(gs)
	}

	for _, game := range games {
		if game.State != realStates[game.ID.String()] {
			game.State = realStates[game.ID.String()]
			game.UpdatedAt = now
		}
	}

	if len(updates) > 0 {
		_, err := tx.Model(&updates).Column("state", "updated_at").Update()
		return err
	}

	return nil
}

func getByNameAndUserId(tx *pg.Tx, game *gamev1Model.Game, name string, userId uuid.UUID) error {
	return tx.Model(game).
		Where("name = ?", name).
		Where("user_id = ?", userId).
		First()
}
