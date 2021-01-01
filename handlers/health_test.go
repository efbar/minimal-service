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

	tt := []struct {
		name        string
		method      string
		contentType string
		path        string
		body        string
		response    string
		status      int
		err         error
	}{
		{
			name:     "GET request",
			method:   "GET",
			path:     "/health",
			response: "Status OK",
			status:   http.StatusOK,
		},
		{
			name:   "POST request",
			method: "POST",
			path:   "/health",
			status: http.StatusMethodNotAllowed,
		},
	}

	for _, tr := range tt {
		req := httptest.NewRequest(tr.method, tr.path, nil)

		rr := httptest.NewRecorder()
		handler := setupHealthTest(t)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, tr.status, rr.Code)
		if tr.method == "GET" && tr.path == "/health" {
			assert.Equal(t, tr.response, rr.Body.String())
		}
	}
}
