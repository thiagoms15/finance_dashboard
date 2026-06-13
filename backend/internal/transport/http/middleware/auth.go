package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/thiago/finance/backend/internal/auth"
)

const userContextKey = "userID"

func JWTAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "unauthorized", "message": "missing bearer token"},
			})
			return
		}

		claims, err := tokens.ParseAccessToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "unauthorized", "message": "invalid token"},
			})
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "unauthorized", "message": "invalid token subject"},
			})
			return
		}

		c.Set(userContextKey, userID)
		c.Next()
	}
}

func UserID(c *gin.Context) uuid.UUID {
	value, ok := c.Get(userContextKey)
	if !ok {
		return uuid.Nil
	}
	userID, _ := value.(uuid.UUID)
	return userID
}
