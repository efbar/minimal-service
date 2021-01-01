package handlers

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
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

	h.l.Info("Delay == %s", strconv.Itoa(delay))
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(delay)

	h.l.Info("Delay == %d, so sleeping %d seconds...\n", strconv.Itoa(n), strconv.Itoa(n))
	time.Sleep(time.Duration(n) * time.Second)
	h.l.Info("Done")

	return err
}
