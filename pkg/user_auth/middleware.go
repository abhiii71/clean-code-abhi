package userauth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the JWT token and sets the claims in the context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "missing or malformed tokens"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims, err := ValidateToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user_uuid and email in context for use in handlers
		c.Set("user_uuid", claims.UserUUID)
		c.Set("email", claims.Email)
		c.Set("age", claims.Age)

		c.Next()
	}
}
