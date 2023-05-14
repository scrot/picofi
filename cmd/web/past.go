package main

import (
	"fmt"
	"net/http"
)

// handlePast guides a person to steps for indexing their past
// with regards to money. It contains net worth and past savings dialogs.
func (s Server) handlePast(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("handleOverview: new request", "method", r.Method, "uri", r.RequestURI, "params", r.URL.RawQuery)

	switch r.Method {
	case http.MethodGet:
		s.newPast(w)
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, OPTIONS")
		s.writeError(w, fmt.Errorf("invalid method for route"), http.StatusMethodNotAllowed)
	}
}

func (s Server) newPast(w http.ResponseWriter) {
	s.writeTemplate(w, "past.tmpl", nil, false)
}
