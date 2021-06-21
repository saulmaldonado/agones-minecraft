package sessions

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	SessionNamev1 string = "agones-minecraft-api-v1"
	UserIDKey     string = "userId"

	SessionPath string = "/api/v1"
	// 7 days
	SessionMaxAge int = 86400 * 7
)

const ()

// Set auth session for request
func SetSession(c *gin.Context, userId uuid.UUID) error {
	sess := sessions.DefaultMany(c, SessionNamev1)

	sess.Set(UserIDKey, userId.String())
	sess.Options(sessions.Options{
		Path:     SessionPath,
		MaxAge:   SessionMaxAge,
		Secure:   true,
		HttpOnly: true,
	})

	return sess.Save()
}

func GetSessionUserId(c *gin.Context) uuid.UUID {
	sess := sessions.DefaultMany(c, SessionNamev1)

	v := sess.Get(UserIDKey)
	userIdString, ok := v.(string)

	if v == nil || !ok {
		return uuid.Nil
	}

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return uuid.Nil
	}

	return userId
}

func DestroySession(c *gin.Context) error {
	sess := sessions.DefaultMany(c, SessionNamev1)
	sess.Clear()
	return sess.Save()
}
