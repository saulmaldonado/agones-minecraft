package session

import (
	"github.com/gin-gonic/gin"

	s "github.com/gin-contrib/sessions"

	sessionsSvc "agones-minecraft/services/auth/sessions"
)

func Sessions() gin.HandlerFunc {
	return s.Sessions(sessionsSvc.SessionNamev1, sessionsSvc.Store)
}
