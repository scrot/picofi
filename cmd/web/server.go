package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/scrot/picofi"
	"github.com/scrot/picofi/cmd/web/templates"
	"golang.org/x/exp/slog"
)

type Server struct {
	logger     *slog.Logger
	tcache     TemplateCache
	tfunctions template.FuncMap
}

func NewServer(logger *slog.Logger, calculator picofi.Calculator) Server {
	server := Server{
		logger:     logger,
		tcache:     make(TemplateCache),
		tfunctions: NewTF(&calculator).FuncMap(),
	}
	return server
}

type TemplateCache = map[string]*template.Template

type CalculatorInput struct {
	AnnualIncome   float64
	AnnualExpenses float64
}

// handleCalculator handles requests for interacting with the calculator
func (s Server) handleCalculator(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("handleCalculator: new request", "method", r.Method, "uri", r.RequestURI, "params", r.URL.RawQuery)

	switch r.Method {
	case http.MethodGet:
		s.newCalculator(w)
	case http.MethodPost:
		s.updateCalculator(w, r)
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		s.writeError(w, fmt.Errorf("handleCalculator: no route for method"), http.StatusMethodNotAllowed)
	}
}

// handleStatic serves static assets like .css and favicons
func (s Server) handleStatic() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.FS(templates.Files)))
}

// newCalculator renders the calculator with the default values
func (s Server) newCalculator(w http.ResponseWriter) {
	data := CalculatorInput{
		AnnualIncome:   70000,
		AnnualExpenses: 50000,
	}

	s.writeTemplate(w, "calculator", data, false)
}

// updateCalculator re-renders the calculator with input provided in the form
func (s Server) updateCalculator(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.writeError(w, fmt.Errorf("updateCalculator: parsing form data: %w", err), http.StatusBadRequest)
	}

	income, err := strconv.ParseFloat(r.Form.Get("income"), 64)
	if err != nil {
		s.writeError(w, fmt.Errorf("updateCalculator: converting form value (income): %w", err), http.StatusBadRequest)
	}

	expenses, err := strconv.ParseFloat(r.Form.Get("expenses"), 64)
	if err != nil {
		s.writeError(w, fmt.Errorf("updateCalculator: converting form value (expenses): %w", err), http.StatusBadRequest)
	}

	data := CalculatorInput{
		AnnualIncome:   income,
		AnnualExpenses: expenses,
	}
	s.writeTemplate(w, "calculator", data, true)
}

// writeTemplate renders and writes .tmpl files to w, exposing functions and data to the template
// template is the template name without the .tmpl extension and will be embedded in base.tmpl
// templates are cached first time they are generated, use rerender when data has been updated
func (s Server) writeTemplate(w http.ResponseWriter, template string, data any, rerender bool) {
	// purge template from cache
	if rerender {
		s.logger.Info(fmt.Sprintf("writeTemplate: re-rendering template %s", template))
		delete(s.tcache, template)
	}

	// render and cache template
	_, ok := s.tcache[template]
	if !ok {
		t, err := renderTemplate(template, s.tfunctions)
		s.logger.Info(fmt.Sprintf("writeTemplate: new template %s cached", template))
		if err != nil {
			s.writeError(w, fmt.Errorf("writeTemplate: render template: %w", err), http.StatusInternalServerError)
		}
		s.tcache[template] = t
	}

	// write template to buffer
	var buf bytes.Buffer
	if err := s.tcache[template].ExecuteTemplate(&buf, "base", data); err != nil {
		s.writeError(w, fmt.Errorf("writeTemplate: execute template: %w", err), http.StatusInternalServerError)
	}

	w.Write(buf.Bytes())
}

func (s Server) writeError(w http.ResponseWriter, err error, code int) {
	s.logger.Error(err.Error())
	http.Error(w, http.StatusText(code), code)
	return
}

// renderTemplate generates a new template.Template from <name>.tmpl and data
// it exposes the FuncMap funcs that can be used in the templates
func renderTemplate(name string, funcs template.FuncMap) (*template.Template, error) {
	files := []string{
		"base.tmpl",
		fmt.Sprintf("%s.tmpl", name),
	}

	t, err := template.New(name).Funcs(funcs).ParseFS(templates.Files, files...)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	return t, nil
}
