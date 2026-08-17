package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"frotago-telemetry-api/config"
	"frotago-telemetry-api/models"
	"frotago-telemetry-api/services"

	"github.com/gin-gonic/gin"
)

// IngestSinglePacket recebe 1 pacote de telemetria GPS/OBD
func IngestSinglePacket(c *gin.Context) {
	var packet models.TelemetryPacket
	if err := c.ShouldBindJSON(&packet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Formato de pacote inválido: " + err.Error()})
		return
	}

	if packet.Timestamp.IsZero() {
		packet.Timestamp = time.Now()
	}

	// Se não veio vehicleId mas veio trackerImei, resolve pelo IMEI
	if packet.VehicleID == "" && packet.TrackerIMEI != "" {
		var vehicle struct{ ID string }
		config.DB.Table("vehicles").Select("id").Where("tracker_imei = ?", packet.TrackerIMEI).Scan(&vehicle)
		if vehicle.ID != "" {
			packet.VehicleID = vehicle.ID
		} else {
			packet.VehicleID = packet.TrackerIMEI
		}
	}

	// 1. Grava no Redis em alta velocidade
	services.SaveLivePositionToRedis(packet)

	// 2. Atualiza a tabela principal `vehicles` no PostgreSQL
	config.DB.Table("vehicles").Where("id = ? OR tracker_imei = ?", packet.VehicleID, packet.TrackerIMEI).Updates(map[string]interface{}{
		"latitude":      packet.Latitude,
		"longitude":     packet.Longitude,
		"speed_km_h":    packet.SpeedKmH,
		"heading":       packet.Heading,
		"fuel_level":    packet.FuelLevel,
		"oil_health":    packet.OilHealth,
		"engine_temp_c": packet.EngineTempC,
		"odometer_km":   packet.OdometerKm,
		"updated_at":    time.Now(),
	})

	// 3. Salva no histórico de logs
	config.DB.Create(&models.TelemetryLog{
		VehicleID:   packet.VehicleID,
		Latitude:    packet.Latitude,
		Longitude:   packet.Longitude,
		SpeedKmH:    packet.SpeedKmH,
		Heading:     packet.Heading,
		FuelLevel:   packet.FuelLevel,
		OilHealth:   packet.OilHealth,
		EngineTempC: packet.EngineTempC,
		OdometerKm:  packet.OdometerKm,
		RecordedAt:  packet.Timestamp,
	})

	// 4. Avalia regras de alerta em segundo plano
	go services.EvaluateTelemetryAlerts(packet)

	// 5. Notifica o WebSockets da Frota
	fleetURL := os.Getenv("FLEET_API_URL")
	if fleetURL == "" {
		fleetURL = "http://localhost:8082"
	}
	go func(p models.TelemetryPacket) {
		url := fmt.Sprintf("%s/api/vehicles/%s/telemetry", fleetURL, p.VehicleID)
		body, _ := json.Marshal(p)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 2 * time.Second}
		_, _ = client.Do(req)
	}(packet)

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"vehicleId": packet.VehicleID,
		"storedAt":  packet.Timestamp,
	})
}

// IngestBatchPackets recebe múltiplos pacotes de uma vez (dispositivos IoT offline sincronizando)
func IngestBatchPackets(c *gin.Context) {
	var packets []models.TelemetryPacket
	if err := c.ShouldBindJSON(&packets); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Array de pacotes inválido"})
		return
	}

	for _, packet := range packets {
		if packet.Timestamp.IsZero() {
			packet.Timestamp = time.Now()
		}
		config.DB.Create(&models.TelemetryLog{
			VehicleID:   packet.VehicleID,
			Latitude:    packet.Latitude,
			Longitude:   packet.Longitude,
			SpeedKmH:    packet.SpeedKmH,
			Heading:     packet.Heading,
			FuelLevel:   packet.FuelLevel,
			OilHealth:   packet.OilHealth,
			EngineTempC: packet.EngineTempC,
			OdometerKm:  packet.OdometerKm,
			RecordedAt:  packet.Timestamp,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"processedCount": len(packets),
	})
}
