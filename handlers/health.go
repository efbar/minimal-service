package handlers

import (
	"fmt"
	"net/http"

	"github.com/efbar/minimal-service/logging"
)

// Health ...
type Health struct {
	log logging.Logger
}

// HandlerHealth ...
func HandlerHealth(l logging.Logger) *Health {
	return &Health{
		log: l,
	}
}

// ServeHTTP ...
func (h *Health) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.log.Info(r.Method, "on", r.URL.String(), "from", r.RemoteAddr)

	if r.Method == http.MethodGet {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprint(rw, "Status OK")
		h.log.Info("Status OK")
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

}
