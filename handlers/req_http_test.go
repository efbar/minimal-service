package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/efbar/minimal-service/helpers"
	"github.com/efbar/minimal-service/logging"
	"github.com/stretchr/testify/assert"
)

func setupReqHTTPTest(t *testing.T) *Data {
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

func TestReqHTTPResp(t *testing.T) {
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
			name:   "GET request on root path",
			method: "GET",
			path:   "/",
			status: http.StatusOK,
			err:    nil,
		},
		{
			name:   "GET request on /test path",
			method: "GET",
			path:   "/test",
			status: http.StatusOK,
			err:    nil,
		},
		{
			name:        "GET request on / path, text plain version",
			method:      "GET",
			contentType: "text/plain",
			path:        "/",
			status:      http.StatusOK,
			response:    "Request served by",
			err:         nil,
		},
		{
			name:        "POST request on /bounce path, text plain version",
			method:      "GET",
			contentType: "text/plain",
			path:        "/",
			status:      http.StatusOK,
			response:    "Request served by",
			err:         nil,
		},
		{
			name:   "POST request on / path",
			method: "POST",
			path:   "/",
			body:   "",
			status: http.StatusMethodNotAllowed,
			err:    nil,
		},
		{
			name:     "POST request on /bounce path, wrong endpoint in body",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:443\"}",
			status:   http.StatusBadGateway,
			response: "Bad Gateway\n",
			err:      nil,
		},
		{
			name:     "POST request on /bounce path, schemeless endpoint in body",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"www.google.it:443\"}",
			status:   http.StatusBadRequest,
			response: "Bad Request\n",
			err:      nil,
		},
		{
			name:     "POST request on /bounce path, schemeless endpoint in body",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"hi/there?\"}",
			status:   http.StatusBadRequest,
			response: "Bad Request\n",
			err:      nil,
		},
		{
			name:     "POST request on /bounce path, endpoint without port",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it\"}",
			status:   http.StatusOK,
			response: "body\n",
			err:      nil,
		},
		{
			name:     "POST request on /bounce path, endpoint with https scheme",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"https://www.google.it\"}",
			status:   http.StatusOK,
			response: "body\n",
			err:      nil,
		},
		{
			name:     "POST request on /bounce path, right endpoint in body",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:80\"}",
			status:   http.StatusOK,
			response: "body",
			err:      nil,
		},
		{
			name:     "POST request on /bounce path, unresolveble DNS",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://fakedns\"}",
			status:   http.StatusBadGateway,
			response: "no such host",
			err:      nil,
		},
		{
			name:     "POST request on /bounce path, unresolveble DNS",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://fakedns\"}",
			status:   http.StatusBadGateway,
			response: "connection refused",
			err:      nil,
		},
	}

	handler := setupReqHTTPTest(t)
	helpers.ListEnvs = map[string]string{
		"DELAY_MAX": "1",
		"TRACING":   "1",
	}

	for _, tr := range tt {

		req := httptest.NewRequest(tr.method, tr.path, strings.NewReader(tr.body))
		if tr.contentType == "text/plain" {
			req.Header.Set("Content-type", "text/plain")
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if tr.method == "GET" && tr.contentType == "application/json" {
			assert.Equal(t, tr.status, rr.Result().StatusCode)
		}

		if tr.method == "GET" && tr.contentType == "text/plain" {
			assert.Equal(t, tr.status, rr.Result().StatusCode)
			assert.Contains(t, rr.Body.String(), tr.response)
		}

		if tr.method == "POST" && tr.path != "/bounce" {

			assert.Equal(t, http.StatusMethodNotAllowed, rr.Result().StatusCode)
		}

		if tr.method == "POST" && tr.path != "/bounce" && tr.body == "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:443\"}" {
			assert.Equal(t, tr.status, rr.Result().StatusCode)
			assert.Equal(t, tr.response, rr.Body.String())
		}

		if tr.method == "POST" && tr.path != "/bounce" && tr.body == "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:80\"}" {

			var tmpl = &JSONResponse{}
			getJBody(rr.Body, tmpl)
			assert.Equal(t, tr.method, tmpl.RequestURI)
			assert.Equal(t, tr.status, rr.Result().StatusCode)
			assert.NotContains(t, rr.Body.String(), tr.response)
		}

		if tr.method == "POST" && tr.path != "/bounce" && tr.body == "{\"rebound\":\"true\",\"endpoint\":\"http://fakedns\"}" {
			assert.Equal(t, tr.status, rr.Result().StatusCode)
			assert.Contains(t, rr.Body.String(), tr.response)
		}

		if tr.method == "POST" && tr.path != "/bounce" && tr.body == "{\"rebound\":\"true\",\"endpoint\":\"http://127.0.0.1:7777\"}" {
			assert.Equal(t, tr.status, rr.Result().StatusCode)
			assert.Contains(t, rr.Body.String(), tr.response)
		}
	}
}

func TestRouting(t *testing.T) {
	handler := setupReqHTTPTest(t)
	server := httptest.NewServer(handler)

	res, err := http.Get(fmt.Sprintf("%s/health", server.URL))
	if err != nil {
		t.Fatalf("GET %v failed", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func getJBody(body *bytes.Buffer, tmpl *JSONResponse) {
	jsonRes := json.NewDecoder(strings.NewReader(body.String()))
	err := jsonRes.Decode(tmpl)
	if err != nil {
		log.Fatalln(err)
	}
}
