package handlers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/efbar/minimal-service/logging"
	"github.com/stretchr/testify/assert"
)

func setupHealthTest(t *testing.T) *Health {
	l := log.New(os.Stdout,
		"Test Logger: ",
		log.Ldate|log.Ltime)
	logger := &logging.Logger{
		Logger: l,
	}
	return &Health{
		*logger,
	}
}

func TestHealthResp(t *testing.T) {

	req := httptest.NewRequest("GET", "/health", nil)

	rr := httptest.NewRecorder()
	handler := setupHealthTest(t)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Status OK", rr.Body.String())
}
