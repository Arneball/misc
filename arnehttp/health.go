package arnehttp

import (
	"net/http"
	"sync/atomic"
)

type Health struct {
	shuttingDown atomic.Bool
}

func (h *Health) Shutdown() {
	h.shuttingDown.Store(true)
}

func (h *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HealthCheck(w, r)
}

func (h *Health) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if h.shuttingDown.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
