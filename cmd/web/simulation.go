package main

import (
	"fmt"
	"net/http"
	"strconv"
)

// handleSimulation guides users for running FI simulations
func (s Server) handleSimulation(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("handleSimulation: new request", "method", r.Method, "uri", r.RequestURI, "params", r.URL.RawQuery)

	switch r.Method {
	case http.MethodGet:
		s.loadSimulation(w, r)
	case http.MethodPost:
		s.updateSimulation(w, r)
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		s.writeError(w, fmt.Errorf("invalid method for route"), http.StatusMethodNotAllowed)
		return
	}
}

// loadSimulation renders the calculator with the default values
func (s Server) loadSimulation(w http.ResponseWriter, r *http.Request) {
	id := s.sessionFromCookie(r)
	session, err := s.loadSession(id)
	if err != nil {
		s.writeError(w, fmt.Errorf("retreiving session from cache: %w", err), http.StatusInternalServerError)
		return
	}
	s.writeTemplate(w, "simulation", session, false)
}

// updateSimulation re-renders the calculator with input provided in the form
func (s Server) updateSimulation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.writeError(w, fmt.Errorf("parsing form data: %w", err), http.StatusBadRequest)
		return
	}

	income, err := strconv.ParseFloat(r.Form.Get("income"), 64)
	if err != nil {
		s.writeError(w, fmt.Errorf("converting form value (income): %w", err), http.StatusBadRequest)
		return
	}

	expenses, err := strconv.ParseFloat(r.Form.Get("expenses"), 64)
	if err != nil {
		s.writeError(w, fmt.Errorf("converting form value (expenses): %w", err), http.StatusBadRequest)
		return
	}

	id := s.sessionFromCookie(r)
	session, err := s.loadSession(id)
	if err != nil {
		s.writeError(w, fmt.Errorf("retreiving session from cache: %w", err), http.StatusInternalServerError)
		return
	}

	session.AnnualIncome = income
	session.AnnualExpenses = expenses

	if err := s.saveSession(id, session); err != nil {
		s.writeError(w, fmt.Errorf("persisting session to cache: %w", err), http.StatusInternalServerError)
		return
	}

	s.writeTemplate(w, "simulation", session, true)
}

func (s Server) updateSavings(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.writeError(w, fmt.Errorf("parsing form data: %w", err), http.StatusBadRequest)
		return
	}

	var (
		i       int
		savings []SavingsRow
	)

	for {
		key := fmt.Sprintf("savings[%d].", i)

		name := r.Form.Get(key + "name")
		if name == "" {
			break
		}

		amount, err := strconv.ParseFloat(r.Form.Get(key+"amount"), 64)
		if err != nil {
			s.writeError(w, fmt.Errorf("converting form value (amount): %w", err), http.StatusBadRequest)
		}

		intrest, err := strconv.ParseFloat(r.Form.Get(key+"intrest"), 64)
		if err != nil {
			s.writeError(w, fmt.Errorf("converting form value (intrest): %w", err), http.StatusBadRequest)
		}

		savings = append(savings, SavingsRow{name, amount, intrest})

		i++
	}

	data := Session{
		Savings: savings,
	}

	s.writeTemplate(w, "simulation", data, true)

}
