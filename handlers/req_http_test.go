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
		envs        map[string]string
	}{
		{
			name:   "GET request on root path",
			method: "GET",
			path:   "/",
			status: http.StatusOK,
		},
		{
			name:   "GET request on root path, delay test",
			method: "GET",
			path:   "/",
			status: http.StatusOK,
			envs: map[string]string{
				"DELAY_MAX": "1",
				"TRACING":   "0",
			},
		},
		{
			name:   "GET request on root path, tracing test",
			method: "GET",
			path:   "/",
			status: http.StatusOK,
			envs: map[string]string{
				"DELAY_MAX": "0",
				"TRACING":   "1",
			},
		},
		{
			name:   "GET request on root path, reject test",
			method: "GET",
			path:   "/",
			status: http.StatusInternalServerError,
			envs: map[string]string{
				"DISCARD_QUOTA": "100",
				"REJECT":        "1",
			},
			response: "Internal Server Error\n",
		},
		{
			name:   "GET request on /test path",
			method: "GET",
			path:   "/test",
			status: http.StatusOK,
		},
		{
			name:        "GET request on / path, text plain version",
			method:      "GET",
			contentType: "text/plain",
			path:        "/",
			status:      http.StatusOK,
			response:    "Request served by",
		},
		{
			name:        "POST request on /bounce path, text plain version",
			method:      "GET",
			contentType: "text/plain",
			path:        "/",
			status:      http.StatusOK,
			response:    "Request served by",
		},
		{
			name:   "POST request on / path",
			method: "POST",
			path:   "/",
			body:   "",
			status: http.StatusMethodNotAllowed,
		},
		{
			name:     "POST request on /bounce path, wrong endpoint in body",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:443\"}",
			status:   http.StatusBadGateway,
			response: "Bad Gateway\n",
		},
		{
			name:     "POST request on /bounce path, schemeless endpoint in body",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"www.google.it:443\"}",
			status:   http.StatusBadGateway,
			response: "Bad Gateway\n",
		},
		{
			name:     "POST request on /bounce path, unresolveble in body",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"hi/there?\"}",
			status:   http.StatusBadGateway,
			response: "Bad Gateway\n",
		},
		{
			name:     "POST request on /bounce path, endpoint without port",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it\"}",
			status:   http.StatusOK,
			response: "Status" + http.StatusText(200),
		},
		{
			name:   "POST request on /bounce path, endpoint with https scheme",
			method: "POST",
			path:   "/bounce",
			body:   "{\"rebound\":\"true\",\"endpoint\":\"https://www.google.it\"}",
			status: http.StatusOK,
		},
		{
			name:     "POST request on /bounce path, right endpoint in body",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:80\"}",
			status:   http.StatusOK,
			response: "200 OK",
		},
		{
			name:     "POST request on /bounce path, unresolveble DNS",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://fakedns\"}",
			status:   http.StatusBadGateway,
			response: "Bad Gateway\n",
		},
		{
			name:     "POST request on /bounce path, unresolveble DNS",
			method:   "POST",
			path:     "/bounce",
			body:     "{\"rebound\":\"true\",\"endpoint\":\"http://127.0.0.1:7777\"}",
			status:   http.StatusBadGateway,
			response: "Bad Gateway\n",
		},
	}

	for _, tr := range tt {
		handler := setupReqHTTPTest(t)
		if len(tr.envs) == 0 {
			handler.envs = map[string]string{
				"DELAY_MAX": "0",
				"TRACING":   "0",
			}
		} else {
			handler.envs = tr.envs
		}
		t.Run(tr.name, func(t *testing.T) {
			req := httptest.NewRequest(tr.method, tr.path, strings.NewReader(tr.body))
			if tr.contentType == "text/plain" {
				req.Header.Set("Content-type", "text/plain")
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tr.status, rr.Result().StatusCode)

			if (tr.method == "GET" && tr.envs["REJECT"] == "1") ||
				(tr.method == "POST" && tr.path == "/bounce" && tr.body == "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:443\"}") {
				assert.Equal(t, tr.response, rr.Body.String())
			}

			if tr.method == "POST" && tr.path == "/bounce" && tr.body == "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:80\"}" {
				var tmpl = &JSONResponse{}
				getJBody(rr.Body, tmpl)

				assert.Equal(t, tr.path, tmpl.RequestURI)
				assert.Equal(t, tr.response, tmpl.Body)
			}

			if (tr.method == "GET" && tr.contentType == "text/plain") ||
				(tr.method == "POST" && tr.path == "/bounce" &&
					(tr.body == "{\"rebound\":\"true\",\"endpoint\":\"http://127.0.0.1:7777\"}" ||
						tr.body == "{\"rebound\":\"true\",\"endpoint\":\"http://fakedns\"}")) {
				assert.Contains(t, rr.Body.String(), tr.response)
			}
		})
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
