package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"frotago-telemetry-api/config"
	"frotago-telemetry-api/handlers"
	"frotago-telemetry-api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ingestAuth, err := middleware.NewDeviceAuthMiddleware(os.Getenv("TELEMETRY_DEVICE_KEYS"))
	if err != nil {
		log.Fatalf("Configuração de autenticação de rastreadores inválida: %v", err)
	}
	if len(os.Getenv("JWT_SECRET")) < 32 {
		log.Fatal("JWT_SECRET deve ter pelo menos 32 caracteres")
	}

	// Conecta com o PostgreSQL (fleet_db) e Redis
	config.ConnectDB()
	config.ConnectRedis()

	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	allowedOrigins := strings.FieldsFunc(os.Getenv("ALLOWED_ORIGINS"), func(r rune) bool {
		return r == ',' || r == ';'
	})
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:5173"}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
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

	ingest := r.Group("/api/telemetry")
	ingest.Use(ingestAuth)
	{
		ingest.POST("/ingest", handlers.IngestSinglePacket)
		ingest.POST("/ingest/batch", handlers.IngestBatchPackets)
	}

	history := r.Group("/api/telemetry")
	history.Use(middleware.JWTAuthMiddleware(), middleware.RequireRole("admin", "member"))
	{
		history.GET("/history/:vehicleId", handlers.GetVehicleHistory)
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
