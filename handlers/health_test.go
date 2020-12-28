package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var logHandle io.Writer

func setupHealthTest(t *testing.T) *Health {
	logHandle := os.Stdout
	return &Health{
		log.New(logHandle,
			"Test Logger: ",
			log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func TestHealthResp(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)

	rr := httptest.NewRecorder()
	handler := setupHealthTest(t)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \"%v\" want \"%v\"",
			status, http.StatusOK)
	}

	expected := `Status OK`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got \"%v\" want \"%v\"",
			rr.Body.String(), expected)
	}

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Status OK", rr.Body.String())
}
