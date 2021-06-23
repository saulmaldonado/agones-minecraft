package user

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"agones-minecraft/config"
	"agones-minecraft/db"
	twitchv1Model "agones-minecraft/models/v1/twitch"
	"agones-minecraft/services/auth/twitch"
)

func CreateTwitchAccount(tx *pg.Tx, account *twitchv1Model.TwitchAccount) error {
	_, err := tx.Model(account).Insert()
	return err
}

func GetTwitchAccountByUserId(tx *pg.Tx, userId uuid.UUID, account *twitchv1Model.TwitchAccount) error {
	return tx.Model(account).Where("user_id = ?", userId).First()
}

func UpdateTwitchAccount(tx *pg.Tx, account *twitchv1Model.TwitchAccount) error {
	account.UpdatedAt = time.Now()
	_, err := tx.Model(account).WherePK().UpdateNotZero()
	return err
}

func UpdateTwitchAccountTokens(tx *pg.Tx, account *twitchv1Model.TwitchAccount) error {
	account.UpdatedAt = time.Now()
	_, err := tx.Model(account).
		Set("access_token = ?access_token", account.AccessToken).
		Set("refresh_token = ?refresh_token", account.RefreshToken).
		Set("updated_at = ?updated_at", account.UpdatedAt).
		WherePK().
		UpdateNotZero()
	return err
}

func RevokeTwitchTokensForUser(tx *pg.Tx, userId uuid.UUID) error {
	var account twitchv1Model.TwitchAccount
	if err := GetTwitchAccountByUserId(tx, userId, &account); err != nil {
		return err
	}

	go RevokeOldTwitchTokens(account)
	return nil
}

func RevokeOldTwitchTokens(tokens twitchv1Model.TwitchAccount) {
	errs := twitch.RevokeTokens(tokens.AccessToken, tokens.RefreshToken, config.GetTwichCreds().ClientID)
	for _, e := range errs {
		zap.L().Warn("error revoking old twitch tokens", zap.Error(e))
	}
}

func ValidateAndRefreshTwitchTokensForUser(userId uuid.UUID) error {
	return db.DB().RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		var account twitchv1Model.TwitchAccount
		if err := GetTwitchAccountByUserId(tx, userId, &account); err != nil {
			return err
		}

		if err := twitch.ValidateToken(account.AccessToken); err != nil {
			if err == twitch.ErrInvalidAccessToken {
				return RefreshTwitchTokensForUser(tx, userId, &account)
			}
		}

		return nil
	})
}

func RefreshTwitchTokensForUser(tx *pg.Tx, userId uuid.UUID, account *twitchv1Model.TwitchAccount) error {
	creds := config.GetTwichCreds()

	newTokens, err := twitch.Refresh(account.RefreshToken, creds.ClientID, creds.ClientSecret)
	if err != nil {
		return err
	}

	account.AccessToken = newTokens.AccessToken
	account.RefreshToken = newTokens.RefreshToken

	return UpdateTwitchAccount(tx, account)
}
