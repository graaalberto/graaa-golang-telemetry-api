package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"frotago-telemetry-api/config"
	"frotago-telemetry-api/models"
	"frotago-telemetry-api/services"

	"github.com/gin-gonic/gin"
)

// IngestSinglePacket recebe 1 pacote de telemetria GPS/OBD
func IngestSinglePacket(c *gin.Context) {
	var packet models.TelemetryPacket
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	if err := c.ShouldBindJSON(&packet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Formato de pacote inválido: " + err.Error()})
		return
	}
	if !authorizePacket(c, &packet) {
		return
	}
	setPacketTimestamp(&packet)

	services.SaveLivePositionToRedis(packet)
	if err := config.DB.Table("vehicles").Where("id = ?", packet.VehicleID).Updates(map[string]interface{}{
		"latitude":      packet.Latitude,
		"longitude":     packet.Longitude,
		"speed_km_h":    packet.SpeedKmH,
		"heading":       packet.Heading,
		"fuel_level":    packet.FuelLevel,
		"oil_health":    packet.OilHealth,
		"engine_temp_c": packet.EngineTempC,
		"odometer_km":   packet.OdometerKm,
		"updated_at":    time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Não foi possível atualizar a viatura"})
		return
	}
	if err := config.DB.Create(&models.TelemetryLog{
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
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Não foi possível guardar a telemetria"})
		return
	}

	go services.EvaluateTelemetryAlerts(packet)

	fleetURL := os.Getenv("FLEET_API_URL")
	if fleetURL == "" {
		fleetURL = "http://localhost:8082"
	}
	go func(p models.TelemetryPacket) {
		serviceToken := os.Getenv("FLEET_TELEMETRY_TOKEN")
		if len(serviceToken) < 32 {
			log.Printf("FLEET_TELEMETRY_TOKEN não configurado; não foi possível notificar fleet-api")
			return
		}
		parsedFleetURL, err := url.Parse(fleetURL)
		if err != nil || parsedFleetURL.Host == "" || parsedFleetURL.User != nil ||
			(parsedFleetURL.Scheme != "https" && !(parsedFleetURL.Scheme == "http" &&
				(parsedFleetURL.Hostname() == "localhost" || parsedFleetURL.Hostname() == "127.0.0.1"))) {
			log.Printf("FLEET_API_URL deve usar HTTPS, exceto em localhost")
			return
		}
		endpoint := fmt.Sprintf("%s/api/vehicles/%s/telemetry", strings.TrimRight(fleetURL, "/"), url.PathEscape(p.VehicleID))
		body, err := json.Marshal(p)
		if err != nil {
			log.Printf("Falha ao serializar telemetria para fleet-api: %v", err)
			return
		}
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Falha ao criar pedido para fleet-api: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Service-Token", serviceToken)
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Falha ao notificar fleet-api: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			log.Printf("fleet-api rejeitou a telemetria com HTTP %d", resp.StatusCode)
			_, _ = io.Copy(io.Discard, resp.Body)
		}
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	if err := c.ShouldBindJSON(&packets); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Array de pacotes inválido"})
		return
	}
	if len(packets) == 0 || len(packets) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "O lote deve conter entre 1 e 1000 pacotes"})
		return
	}

	for _, packet := range packets {
		if !authorizePacket(c, &packet) {
			return
		}
		setPacketTimestamp(&packet)
		if err := config.DB.Create(&models.TelemetryLog{
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
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Não foi possível guardar o lote de telemetria"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"processedCount": len(packets),
	})
}

func authorizePacket(c *gin.Context, packet *models.TelemetryPacket) bool {
	authenticatedIMEI := c.GetString("trackerIMEI")
	if authenticatedIMEI == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Rastreador não autenticado"})
		return false
	}
	if packet.TrackerIMEI != "" && packet.TrackerIMEI != authenticatedIMEI {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "O pacote não pertence ao rastreador autenticado"})
		return false
	}
	packet.TrackerIMEI = authenticatedIMEI

	var vehicle struct {
		ID string
	}
	if err := config.DB.Table("vehicles").Select("id").Where("tracker_imei = ?", authenticatedIMEI).Take(&vehicle).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Rastreador não associado a uma viatura"})
		return false
	}
	if packet.VehicleID != "" && packet.VehicleID != vehicle.ID {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "O rastreador não está autorizado para esta viatura"})
		return false
	}
	packet.VehicleID = vehicle.ID
	if packet.Latitude < -90 || packet.Latitude > 90 || packet.Longitude < -180 || packet.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Coordenadas GPS inválidas"})
		return false
	}
	return true
}

func setPacketTimestamp(packet *models.TelemetryPacket) {
	if packet.Timestamp.IsZero() {
		packet.Timestamp = time.Now()
	}
}
