package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"frotago-telemetry-api/models"
)

// EvaluateTelemetryAlerts verifica se o veículo violou regras críticas
func EvaluateTelemetryAlerts(packet models.TelemetryPacket) {
	var alertType string
	var alertMsg string

	if packet.SpeedKmH > 90.0 {
		alertType = "EXCESSO_VELOCIDADE"
		alertMsg = fmt.Sprintf("Viatura %s detectada a %.1f km/h (Limite: 80 km/h)", packet.VehicleID, packet.SpeedKmH)
	} else if packet.FuelLevel > 0 && packet.FuelLevel < 15.0 {
		alertType = "COMBUSTIVEL_CRITICO"
		alertMsg = fmt.Sprintf("Viatura %s com nível de combustível perigoso: %.1f%%", packet.VehicleID, packet.FuelLevel)
	} else if packet.EngineTempC > 98.0 {
		alertType = "SUPERAQUECIMENTO_MOTOR"
		alertMsg = fmt.Sprintf("Motor da viatura %s atingiu %.1f°C! Risco de avaria mecânica.", packet.VehicleID, packet.EngineTempC)
	}

	if alertType != "" {
		log.Printf("🚨 ALERTA CRÍTICO IOT [%s]: %s", alertType, alertMsg)
		notifyFleetCore(packet.VehicleID, alertType, alertMsg)
	}
}

func notifyFleetCore(vehicleID, alertType, message string) {
	fleetURL := os.Getenv("FLEET_API_URL")
	if fleetURL == "" {
		fleetURL = "http://localhost:8082"
	}

	go func() {
		endpoint := fmt.Sprintf("%s/api/breakdowns", fleetURL)
		body, _ := json.Marshal(map[string]interface{}{
			"vehicleId":   vehicleID,
			"severity":    "high",
			"category":    "electrical",
			"description": fmt.Sprintf("[Alerta Automático IoT - %s] %s", alertType, message),
		})

		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 3 * time.Second}
			_, _ = client.Do(req)
		}
	}()
}
