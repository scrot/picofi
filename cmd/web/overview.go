package main

import (
	"fmt"
	"net/http"
)

// handleOverview provides a consolidated view of a person's personal situation
// and provides a introduction to PicoFI for newcomers
func (s Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("handleOverview: new request", "method", r.Method, "uri", r.RequestURI, "params", r.URL.RawQuery)

	switch r.Method {
	case http.MethodGet:
		s.newOverview(w)
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, OPTIONS")
		s.writeError(w, fmt.Errorf("invalid method for route"), http.StatusMethodNotAllowed)
	}
}

func (s Server) newOverview(w http.ResponseWriter) {
	s.writeTemplate(w, "overview", nil, false)
}
