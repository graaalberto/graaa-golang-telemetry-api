package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"frotago-telemetry-api/config"
	"frotago-telemetry-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Conecta com o PostgreSQL (fleet_db) e Redis
	config.ConnectDB()
	config.ConnectRedis()

	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "graaa-golang-telemetry-api",
			"database":  "fleet_db",
			"cache":     "redis",
			"version":   "v1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	api := r.Group("/api/telemetry")
	{
		// Rotas de Ingestão IoT
		api.POST("/ingest", handlers.IngestSinglePacket)
		api.POST("/ingest/batch", handlers.IngestBatchPackets)

		// Histórico e Trajetos
		api.GET("/history/:vehicleId", handlers.GetVehicleHistory)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("📡 FrotaGo Telemetry & IoT Ingestion API rodando na porta :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar servidor de telemetria: %v", err)
	}
}
