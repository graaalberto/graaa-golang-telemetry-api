package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const minimumJWTSecretLength = 32

type accessClaims struct {
	UserID    string   `json:"user_id"`
	TokenType string   `json:"token_type"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := os.Getenv("JWT_SECRET")
		if len(secret) < minimumJWTSecretLength {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "Autenticação não configurada"})
			return
		}
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Token de autenticação não fornecido"})
			return
		}

		claims := &accessClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("método de assinatura inválido")
			}
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
		if err != nil || token == nil || !token.Valid || claims.UserID == "" || claims.TokenType != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Token JWT inválido ou expirado"})
			return
		}

		role := "viewer"
		for _, tokenRole := range claims.Roles {
			switch strings.ToLower(tokenRole) {
			case "admin":
				role = "admin"
			case "member":
				if role != "admin" {
					role = "member"
				}
			}
		}
		c.Set("userID", claims.UserID)
		c.Set("userRole", role)
		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("userRole")
		userRole, ok := role.(string)
		if ok {
			for _, allowed := range allowedRoles {
				if strings.EqualFold(userRole, allowed) {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Acesso negado"})
	}
}
