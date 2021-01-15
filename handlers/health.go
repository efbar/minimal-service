package handlers

import (
	"fmt"
	"net/http"

	"github.com/efbar/minimal-service/logging"
)

// Health ...
type Health struct {
	log  logging.Logger
	envs map[string]string
}

// HandlerHealth ...
func HandlerHealth(l logging.Logger, envs map[string]string) *Health {
	return &Health{
		log:  l,
		envs: envs,
	}
}

// ServeHTTP ...
func (h *Health) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.log.Debug(h.envs["DEBUG"], r.Method, "on", r.URL.String(), "from", r.RemoteAddr)

	if r.Method == http.MethodGet {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprint(rw, "Status OK")
		h.log.Debug(h.envs["DEBUG"], "Status OK")
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

}
