package handlers

import (
	"net/http"

	"frotago-telemetry-api/config"
	"frotago-telemetry-api/models"

	"github.com/gin-gonic/gin"
)

// GetVehicleHistory retorna o histórico de posições GPS para desenhar o trajeto no mapa
func GetVehicleHistory(c *gin.Context) {
	vehicleID := c.Param("vehicleId")
	limit := 100

	var logs []models.TelemetryLog
	if err := config.DB.Where("vehicle_id = ?", vehicleID).
		Order("recorded_at desc").
		Limit(limit).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Erro ao buscar histórico"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"vehicleId": vehicleID,
		"count":     len(logs),
		"history":   logs,
	})
}
