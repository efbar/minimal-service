package utils

import (
	"net/http"
)

// GetHeaders ...
func GetHeaders(r *http.Request, m map[string]string) map[string]string {
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
