package handlers

import (
	"fmt"
	"log"
	"net/http"
)

// Health ...
type Health struct {
	l *log.Logger
}

// HandlerHealth ...
func HandlerHealth(l *log.Logger) *Health {
	return &Health{
		l,
	}
}

// ServeHTTP ...
func (h *Health) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.l.Printf("%s on %s from %s", r.Method, r.URL, r.RemoteAddr)

	if r.Method == http.MethodGet {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprint(rw, "Status OK")
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

}
