package server

import (
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/internal/services"

	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
	"time"
)

func AuthMiddleware(errLogger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Set(consts.KeyId, uint(0))
			c.Set(consts.KeyRole, models.RoleAnonymous)
			c.Next()
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			errLogger.Printf("%s", "invalid authorization header format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}
		data, err := jwt.ParseWithClaims(headerParts[1], &services.UserClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(viper.GetString("auth_access_signing_key")), nil
			})
		if data == nil {
			errLogger.Printf("%s", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		claims, ok := data.Claims.(*services.UserClaims)
		if err != nil {
			if claims.ExpiresAt.Unix() < time.Now().Unix() {
				errLogger.Printf("%s", err.Error())
				c.JSON(http.StatusUnauthorized, gin.H{"error": consts.ErrTokenExpired})
				c.Abort()
				return
			}
			errLogger.Printf("%s", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		if !ok {
			errLogger.Printf("%s", consts.ErrNotStandardToken)
			c.JSON(http.StatusUnauthorized, gin.H{"error": consts.ErrNotStandardToken})
			c.Abort()
			return
		}

		c.Set(consts.KeyId, claims.Id)
		c.Set(consts.KeyRole, claims.Role)
		c.Next()
	}
}
