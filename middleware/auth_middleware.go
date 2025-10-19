package middleware

import (
	"net/http"

	"github.com/Neph-dev/MovieStreamServer/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(_context *gin.Context) {
		token, err := utils.GetAccessToken(_context)

		if err != nil {
			_context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
			_context.Abort()
			return
		}

		if token == "" {
			_context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided"})
			_context.Abort()
			return
		}

		claims, err := utils.ValidateToken(token)

		if err != nil {
			_context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
			_context.Abort()
			return
		}

		_context.Set("email", claims.Email)
		_context.Set("first_name", claims.FirstName)
		_context.Set("last_name", claims.LastName)
		_context.Set("userId", claims.UID)
		_context.Set("role", claims.Role)

		_context.Next()
	}
}