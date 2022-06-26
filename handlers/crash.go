package handlers

import (
	"net/http"
	"os"

	"github.com/efbar/minimal-service/logging"
)

// Crash ...
type Crash struct {
	log  logging.Logger
	envs map[string]string
}

// HandlerCrash ...
func HandlerCrash(l logging.Logger, envs map[string]string) *Crash {
	return &Crash{
		log:  l,
		envs: envs,
	}
}

// ServeHTTP ...
func (h *Crash) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.log.Debug(h.envs["DEBUG"], r.Method, "on", r.URL.String(), "from", r.RemoteAddr)

	if r.Method == http.MethodGet {
		h.log.Debug(h.envs["DEBUG"], "Crashing")
		os.Exit(137)
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

}
