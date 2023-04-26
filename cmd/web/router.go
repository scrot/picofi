package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/scrot/picofi/cmd/web/templates"
	"golang.org/x/exp/slog"
)

type Router struct {
	Logger *slog.Logger
}

func (router Router) handleRoot(w http.ResponseWriter, r *http.Request) {
	router.Logger.Info("handleRoot: new request", "method", r.Method, "uri", r.RequestURI, "params", r.URL.RawQuery)

	switch r.Method {
	case http.MethodGet:
		if err := renderTemplate(w, "calculator", nil); err != nil {
			router.Logger.Error("rendering calculator template", "err", err)
		}
	case http.MethodPost:
		//todo
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	default:
		router.Logger.Info("handleRoot: no route for method", "method", r.Method)
		code := http.StatusMethodNotAllowed
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, http.StatusText(code), code)
	}

}

func (router Router) handleStatic() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.FS(templates.Files)))
}

func renderTemplate(w io.Writer, name string, data any) error {
	files := []string{
		"base.tmpl",
		fmt.Sprintf("%s.tmpl", name),
	}

	t, err := template.ParseFS(templates.Files, files...)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
