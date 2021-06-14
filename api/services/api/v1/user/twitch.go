package user

import (
	"github.com/google/uuid"
	"go.uber.org/zap"

	"agones-minecraft/config"
	"agones-minecraft/db"
	twitchv1Model "agones-minecraft/models/v1/twitch"
	"agones-minecraft/services/auth/twitch"
)

// Get TwitchTokens by its user_id foreign key
func GetUserTwitchTokens(userId uuid.UUID, twitchToken *twitchv1Model.TwitchToken) error {
	return db.DB().Where("user_id = ?", userId).First(twitchToken).Error
}

func UpdateUserTwitchTokens(userId uuid.UUID, twitchToken *twitchv1Model.TwitchToken) error {
	return db.DB().Model(twitchv1Model.TwitchToken{}).Where("user_id = ?", userId).Updates(twitchToken).Error
}

func RevokeTwitchTokensForUser(userId uuid.UUID) error {
	var tokens *twitchv1Model.TwitchToken
	if err := GetUserTwitchTokens(userId, tokens); err != nil {
		return err
	}

	go RevokeOldTwitchTokens(tokens)
	return nil
}

func RevokeOldTwitchTokens(tokens *twitchv1Model.TwitchToken) {
	errs := twitch.RevokeTokens(*tokens.TwitchAccessToken, *tokens.TwitchRefreshToken, config.GetTwichCreds().ClientID)
	for _, e := range errs {
		zap.L().Warn("error revoking old twitch tokens", zap.Error(e))
	}
}

func ValidateAndRefreshTwitchTokensForUser(userId uuid.UUID) error {
	var tokens twitchv1Model.TwitchToken
	if err := GetUserTwitchTokens(userId, &tokens); err != nil {
		return err
	}

	if err := twitch.ValidateToken(*tokens.TwitchAccessToken); err != nil {
		if err == twitch.ErrInvalidAccessToken {
			return RefreshTwitchTokensForUser(userId, &tokens)
		}
	}

	return nil
}

func RefreshTwitchTokensForUser(userId uuid.UUID, tokens *twitchv1Model.TwitchToken) error {
	creds := config.GetTwichCreds()

	newTokens, err := twitch.Refresh(*tokens.TwitchRefreshToken, creds.ClientID, creds.ClientSecret)
	if err != nil {
		return err
	}

	tokens.TwitchAccessToken = &newTokens.AccessToken
	tokens.TwitchRefreshToken = &newTokens.RefreshToken

	return UpdateUserTwitchTokens(userId, tokens)
}
