package models

import "time"

// TelemetryPacket é o pacote recebido dos rastreadores GPS / App do Motorista
type TelemetryPacket struct {
	TrackerIMEI string    `json:"trackerImei"`
	VehicleID   string    `json:"vehicleId"`
	Latitude    float64   `json:"lat" binding:"required"`
	Longitude   float64   `json:"lng" binding:"required"`
	SpeedKmH    float64   `json:"speedKmH"`
	Heading     float64   `json:"heading"`
	AltitudeM   float64   `json:"altitudeM"`
	FuelLevel   float64   `json:"fuelLevel"`   // 0 a 100%
	OilHealth   float64   `json:"oilHealth"`   // 0 a 100%
	EngineTempC float64   `json:"engineTempC"` // Celsius
	OdometerKm  float64   `json:"odometerKm"`
	Ignition    bool      `json:"ignition"`
	Timestamp   time.Time `json:"timestamp"`
}

// TelemetryLog é a tabela de histórico no PostgreSQL
type TelemetryLog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID   string    `gorm:"not null;type:varchar(64)" json:"vehicleId"`
	Latitude    float64   `gorm:"not null;type:numeric(10,7)" json:"latitude"`
	Longitude   float64   `gorm:"not null;type:numeric(10,7)" json:"longitude"`
	SpeedKmH    float64   `json:"speedKmH"`
	Heading     float64   `json:"heading"`
	FuelLevel   float64   `json:"fuelLevel"`
	OilHealth   float64   `json:"oilHealth"`
	EngineTempC float64   `json:"engineTempC"`
	OdometerKm  float64   `json:"odometerKm"`
	RecordedAt  time.Time `json:"recordedAt"`
}

// TelemetryAlertRule define os limiares de segurança
type TelemetryAlertRule struct {
	ID                string  `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name              string  `gorm:"not null;type:varchar(100)" json:"name"`
	MetricType        string  `gorm:"not null;type:varchar(50)" json:"metricType"` // fuel_low, engine_overheat, speed_limit
	ThresholdValue    float64 `gorm:"not null" json:"thresholdValue"`
	ConditionOperator string  `gorm:"not null;type:varchar(5)" json:"conditionOperator"` // <, >, <=
	IsActive          bool    `gorm:"default:true" json:"isActive"`
}
