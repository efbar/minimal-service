package handlers

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/efbar/minimal-service/helpers"
	"github.com/efbar/minimal-service/logging"
	"github.com/stretchr/testify/assert"
)

func setupCrashTest(t *testing.T) *Data {
	l := log.New(os.Stdout,
		"Test Logger: ",
		log.Ldate|log.Ltime)
	logger := &logging.Logger{
		Logger: l,
	}

	return &Data{
		*logger,
		helpers.ListEnvs,
	}
}

func TestCrashResp(t *testing.T) {

	tt := []struct {
		name          string
		method        string
		path          string
		contentLength int64
	}{
		{
			name:          "GET request",
			method:        "GET",
			path:          "/crash",
			contentLength: -1,
		},
	}

	for _, tr := range tt {
		req := httptest.NewRequest(tr.method, tr.path, nil)

		rr := httptest.NewRecorder()
		handler := setupCrashTest(t)

		handler.ServeHTTP(rr, req)

		if tr.method == "GET" && tr.path == "/crash" {
			assert.Equal(t, tr.contentLength, rr.Result().ContentLength)
		}
	}
}
