package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const minimumDeviceKeyLength = 32

func NewDeviceAuthMiddleware(configuredKeys string) (gin.HandlerFunc, error) {
	var keys map[string]string
	if err := json.Unmarshal([]byte(configuredKeys), &keys); err != nil || len(keys) == 0 {
		return nil, errors.New("TELEMETRY_DEVICE_KEYS deve ser um objeto JSON não vazio")
	}
	for imei, key := range keys {
		if strings.TrimSpace(imei) == "" || len(key) < minimumDeviceKeyLength {
			return nil, errors.New("cada rastreador deve ter IMEI e chave com pelo menos 32 caracteres")
		}
	}

	return func(c *gin.Context) {
		imei := strings.TrimSpace(c.GetHeader("X-Tracker-IMEI"))
		provided := c.GetHeader("X-Tracker-Token")
		expected, ok := keys[imei]
		if !ok || len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Credencial de rastreador inválida"})
			return
		}
		c.Set("trackerIMEI", imei)
		c.Next()
	}, nil
}
