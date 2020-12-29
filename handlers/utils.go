package handlers

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

// EncodeJSON ...
func (j *JSONResponse) EncodeJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(j)
}

// DecodeJSON ...
func DecodeJSON(body []byte, jp *JSONPost, rw http.ResponseWriter) error {
	err := json.Unmarshal(body, &jp)
	if err != nil {
		http.Error(rw, "Bad Request", http.StatusBadRequest)
	}
	return err
}

// GetHostname ...
func (h *Data) GetHostname() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		h.l.Printf("Server hostname unknown: %s\n\n", err.Error())
	}
	return host, err
}

// CollectHeaders ...
func CollectHeaders(r *http.Request, m map[string]string) map[string]string {
	var headers = map[string]string{}
	for key, values := range r.Header {
		for _, value := range values {
			headers[string(key)] = string(value)
		}
	}
	for key, value := range m {
		headers[key] = string(value)
	}

	return headers
}

// Delayer ...
func (h *Data) Delayer(delayEnv string) error {

	delay, err := strconv.Atoi(delayEnv)

	h.l.Printf("Delay == %d", delay)
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(delay)

	h.l.Printf("Delay == %d, so sleeping %d seconds...\n", n, n)
	time.Sleep(time.Duration(n) * time.Second)
	h.l.Println("Done")

	return err
}
