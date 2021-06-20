package user

import (
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
	_, err := db.DB().Model(account).Insert()
	return err
}

func GetTwitchAccountByUserId(userId uuid.UUID, account *twitchv1Model.TwitchAccount) error {
	return db.DB().Model(account).Where("user_id = ?", userId).First()
}

func UpdateTwitchAccount(account *twitchv1Model.TwitchAccount) error {
	account.UpdatedAt = time.Now()
	_, err := db.DB().Model(account).WherePK().Update()
	return err
}

func RevokeTwitchTokensForUser(userId uuid.UUID) error {
	var account twitchv1Model.TwitchAccount
	if err := GetTwitchAccountByUserId(userId, &account); err != nil {
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
	var account twitchv1Model.TwitchAccount
	if err := GetTwitchAccountByUserId(userId, &account); err != nil {
		return err
	}

	if err := twitch.ValidateToken(account.AccessToken); err != nil {
		if err == twitch.ErrInvalidAccessToken {
			return RefreshTwitchTokensForUser(userId, &account)
		}
	}

	return nil
}

func RefreshTwitchTokensForUser(userId uuid.UUID, account *twitchv1Model.TwitchAccount) error {
	creds := config.GetTwichCreds()

	newTokens, err := twitch.Refresh(account.RefreshToken, creds.ClientID, creds.ClientSecret)
	if err != nil {
		return err
	}

	account.AccessToken = newTokens.AccessToken
	account.RefreshToken = newTokens.RefreshToken

	return UpdateTwitchAccount(account)
}
