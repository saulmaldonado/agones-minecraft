package errors

type ErrorID string

const (
	ErrUnknownID ErrorID = "0000"

	ErrMalformedJSON ErrorID = "0001"

	ErrMissingRequestBody ErrorID = "0002"

	// random bytes read io err
	ErrNewState ErrorID = "b7b3"
	// state cookie encryption or encoding err
	ErrEncodingCookie ErrorID = "f22c"
	// missing state from request
	ErrMissingState ErrorID = "528b"
	// missing state challenge from store
	ErrMissingStateChallenge ErrorID = "59ad"
	// failed state challenge
	ErrFailedStateChallenge ErrorID = "4775"
	// invalid or missing session cookie in request
	ErrUnautorizedSession ErrorID = "1494"
	// error saving session and adding cookie to request
	ErrSavingSession ErrorID = "85a6"
	// error exchanging token for twitch token
	ErrTwitchTokenExchange ErrorID = "56a5"
	// errors getting token payload
	ErrTwitchTokenPayload ErrorID = "bbf8"
	// unverified twitch email
	ErrTwitchUnverifiedEmail ErrorID = "dde8"
	// error updating user in database
	ErrUpdatingUser ErrorID = "2545"
	// error generating new jwt tokens
	ErrGeneratingNewTokens ErrorID = "623a"
	// error saving new jwt tokens to store
	ErrSavingNewTokens ErrorID = "c4f4"
	// error getting jwt token from store
	ErrRetrievingTokens ErrorID = "c26e"
	// error deleting jwt token from store
	ErrDeletingTokens ErrorID = "d266"
	// user id ref missing from request context
	ErrMissingUserId ErrorID = "7149"
	// user missing from database
	ErrUserNotFound ErrorID = "c492"
	// error finding user from database
	ErrRetrievingUser ErrorID = "6a53"
	// edit user validation error
	ErrEditUserValidation ErrorID = "38a2"
	// mc username not found error
	ErrMcUserNotFound ErrorID = "33e3"
	// error unmarshalling mc user json
	ErrUnmarshalingMCAccountJSON = "7808"
	// error getting minecraft user
	ErrRetrievingMcUser ErrorID = "dda6"
	// missing refresh token from request
	ErrMissingRefreshToken ErrorID = "8578"
	// error parsing refresh token
	ErrRefreshTokenParsing ErrorID = "c0cf"
	// expected refresh token
	ErrRefreshTokenExpected ErrorID = "07d3"
	// invalid refresh token
	ErrInvalidRefreshToken ErrorID = "c22e"
	// unable to verify refresh token
	ErrUnableToVerifyRefreshToken ErrorID = "5d2d"
	// missing access token from request
	ErrMissingAccessToken ErrorID = "fa67"
	// error parsing Access token
	ErrAccessTokenParsing ErrorID = "1002"
	// expected Access token
	ErrAccessTokenExpected ErrorID = "87a8"
	// invalid Access token
	ErrInvalidAccessToken ErrorID = "eeb4"
	// unable to verify Access token
	ErrUnableToVerifyAccessToken ErrorID = "4e25"
	// twitch credentials not found in database
	ErrTwitchCredentialsNotFound ErrorID = "6315"
	// error retrieving twitch credentials
	ErrRetrievingTwitchCredentials ErrorID = "0ef9"
	// error validating twitch tokens
	ErrValidatingTwitchToken ErrorID = "25fe"
	// invalid twitch refresh token
	ErrTwitchCredentialsInvalid ErrorID = "7dc5"
	// error listing games from informer
	ErrListingGames ErrorID = "756a"
	// error game not found
	ErrGameNotFound ErrorID = "995d"
	// error retrieving game server
	ErrRetrievingGameServer ErrorID = "86ec"
	// error retrieving game server from db
	ErrRetrievingGameServerFromDB ErrorID = "acd1"
	// create game server validation error
	ErrCreateGameServerValidation ErrorID = "6163"
	// subdomain taken error
	ErrSubdomainTaken ErrorID = "710c"
	// game server name taken
	ErrGameServerNameTaken ErrorID = "dbec"
	// error creating game
	ErrCreatingGame ErrorID = "8410"
	// error deleting game from k8s
	ErrDeletingGameFromK8s ErrorID = "8662"
	// error deleting game from DB
	ErrDeletingGameFromDB ErrorID = "cd6d"
)
