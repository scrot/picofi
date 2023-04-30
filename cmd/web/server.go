package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Rhymond/go-money"
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
	fs := template.FuncMap{
		"annualSaveRate": func(income, expenses float64) string {
			res := calculator.AnnualSaveRate(
				money.NewFromFloat(income, calculator.Currency),
				money.NewFromFloat(expenses, calculator.Currency),
			)
			return res.Display()
		},
	}

	server := Server{
		logger:     logger,
		tcache:     make(TemplateCache),
		tfunctions: fs,
	}

	return server
}

type TemplateCache = map[string]*template.Template

type CalculatorInput struct {
	AnnualIncome   float64
	AnnualExpenses float64
}

func (s Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("handleRoot: new request", "method", r.Method, "uri", r.RequestURI, "params", r.URL.RawQuery)

	switch r.Method {
	case http.MethodGet:
		s.newCalculator(w)
	case http.MethodPost:
		//TODO: handle form input
		s.updateCalculator(w, r)
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	default:
		s.logger.Info("handleRoot: no route for method", "method", r.Method)
		code := http.StatusMethodNotAllowed
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, http.StatusText(code), code)
	}

}

func (s Server) handleStatic() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.FS(templates.Files)))
}

func (s Server) newCalculator(w http.ResponseWriter) {
	data := CalculatorInput{
		AnnualIncome:   70000,
		AnnualExpenses: 50000,
	}

	s.writeTemplate(w, "calculator", data, false)
}

func (s Server) updateCalculator(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.logger.Error("update calculator", "err", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	income, err := strconv.ParseFloat(r.Form.Get("income"), 64)
	if err != nil {
		s.logger.Error("parsing form income value", "err", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	expenses, err := strconv.ParseFloat(r.Form.Get("expenses"), 64)
	if err != nil {
		s.logger.Error("parsing form expenses value", "err", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
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
	// purge template form cache if rerender
	if rerender {
		s.logger.Info(fmt.Sprintf("forced rerender, purge template %s from cache", template))
		delete(s.tcache, template)
	}

	// cache new rendered template
	_, ok := s.tcache[template]
	if !ok {
		s.logger.Info(fmt.Sprintf("template %s not found in cache", template))

		t, err := renderTemplate(template, s.tfunctions)
		s.logger.Info(fmt.Sprintf("template %s cached", template))
		if err != nil {
			s.logger.Error(fmt.Sprintf("writeTemplate: render template %s", template), "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		s.tcache[template] = t
	}

	var buf bytes.Buffer
	if err := s.tcache[template].ExecuteTemplate(&buf, "base", data); err != nil {
		s.logger.Error(fmt.Sprintf("writeTemplate: execute template %s", template), "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes())
	s.logger.Info(fmt.Sprintf("template %s retreived from cache", template))
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
