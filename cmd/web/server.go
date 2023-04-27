package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/Rhymond/go-money"
	"github.com/scrot/picofi"
	"github.com/scrot/picofi/cmd/web/templates"
	"golang.org/x/exp/slog"
)

type Server struct {
	logger     *slog.Logger
	tcache     TemplateCache
	tfunctions template.FuncMap
	tdata      TemplateData
}

func NewServer(logger *slog.Logger, calculator picofi.Calculator) Server {
	data := make(TemplateData)
	data["calculator"] = DefaultCalculatorData

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
		tdata:      data,
	}

	return server
}

type TemplateCache = map[string]*template.Template

type TemplateData = map[string]any

type CalculatorData struct {
	AnnualIncome   float64
	AnnualExpenses float64
}

var DefaultCalculatorData = CalculatorData{
	AnnualIncome:   70000,
	AnnualExpenses: 50000,
}

func (s Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("handleRoot: new request", "method", r.Method, "uri", r.RequestURI, "params", r.URL.RawQuery)

	switch r.Method {
	case http.MethodGet:
		s.writeTemplate(w, "calculator", false)
	case http.MethodPost:
		//TODO: handle form input
		s.writeTemplate(w, "calculator", true)
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

// writeTemplate renders and writes .tmpl files to w, exposing functions and data to the template
// template is the template name without the .tmpl extension and will be embedded in base.tmpl
// templates are cached first time they are generated, use rerender when data has been updated
func (s Server) writeTemplate(w io.Writer, template string, rerender bool) {
	if rerender {
		s.logger.Info(fmt.Sprintf("forced rerender, purge template %s from cache", template))
		delete(s.tcache, template)
	}

	t, ok := s.tcache[template]
	if !ok {
		s.logger.Info(fmt.Sprintf("template %s not found in cache", template))

		t, err := renderTemplate(template, s.tfunctions)
		if err != nil {
			s.logger.Error(fmt.Sprintf("writeTemplate: render template %s", template), "err", err)
		}
		s.tcache[template] = t
	}

	if err := t.ExecuteTemplate(w, "base", s.tdata[template]); err != nil {
		s.logger.Error(fmt.Sprintf("writeTemplate: execute template %s", template), "err", err)
	}
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
