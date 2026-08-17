package services

import (
	"encoding/json"
	"fmt"
	"time"

	"frotago-telemetry-api/config"
	"frotago-telemetry-api/models"

	"github.com/redis/go-redis/v9"
)

// SaveLivePositionToRedis grava a última posição do veículo no cache com expiração de 24h
func SaveLivePositionToRedis(packet models.TelemetryPacket) {
	if config.RedisClient == nil {
		return
	}

	key := fmt.Sprintf("vehicle:live:%s", packet.VehicleID)
	data, err := json.Marshal(packet)
	if err != nil {
		return
	}

	config.RedisClient.Set(config.Ctx, key, data, 24*time.Hour)

	// Adiciona na estrutura Geo do Redis usando &redis.GeoLocation
	config.RedisClient.GeoAdd(config.Ctx, "fleet:geo:locations", &redis.GeoLocation{
		Name:      packet.VehicleID,
		Longitude: packet.Longitude,
		Latitude:  packet.Latitude,
	})
}
