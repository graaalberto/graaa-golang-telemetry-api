package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDeviceAuthMiddlewareBindsKeyToIMEI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := "tracker-secret-with-at-least-thirty-two-bytes"
	auth, err := NewDeviceAuthMiddleware(`{"imei-1":"` + key + `"}`)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.POST("/ingest", auth, func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("trackerIMEI"))
	})

	request := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	request.Header.Set("X-Tracker-IMEI", "imei-1")
	request.Header.Set("X-Tracker-Token", key)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "imei-1" {
		t.Fatalf("valid device status/body = %d/%q", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/ingest", nil)
	request.Header.Set("X-Tracker-IMEI", "imei-2")
	request.Header.Set("X-Tracker-Token", key)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unknown device status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestDeviceAuthMiddlewareRejectsWeakConfiguration(t *testing.T) {
	if _, err := NewDeviceAuthMiddleware(`{"imei-1":"short"}`); err == nil {
		t.Fatal("expected weak device key configuration to be rejected")
	}
}
