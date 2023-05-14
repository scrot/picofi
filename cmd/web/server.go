package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/scrot/picofi"
	"github.com/scrot/picofi/cmd/web/templates"
	"golang.org/x/exp/slog"
)

type Server struct {
	logger     *slog.Logger
	sessions   *bigcache.BigCache
	tcache     TemplateCache
	tfunctions template.FuncMap
}

func NewServer(logger *slog.Logger, calculator picofi.Calculator) Server {
	c, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Hour*24))
	if err != nil {
		logger.Error("unable to create cache", "err", err)
		os.Exit(1)
	}

	server := Server{
		logger:     logger,
		sessions:   c,
		tcache:     make(TemplateCache),
		tfunctions: NewTF(&calculator).FuncMap(),
	}
	return server
}

type TemplateCache = map[string]*template.Template

// handleStatic serves static assets like .css and favicons
func (s Server) handleStatic() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.FS(templates.Files)))
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
			return
		}
		s.tcache[template] = t
	}

	// write template to buffer
	var buf bytes.Buffer
	if err := s.tcache[template].ExecuteTemplate(&buf, "base", data); err != nil {
		s.writeError(w, fmt.Errorf("writeTemplate: execute template: %w", err), http.StatusInternalServerError)
		return
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
		"nav.tmpl",
		fmt.Sprintf("%s.tmpl", name),
	}

	t, err := template.New(name).Funcs(funcs).ParseFS(templates.Files, files...)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	return t, nil
}

func (s Server) sessionFromCookie(r *http.Request) string {
	v, err := r.Cookie(sessionCookieKey)
	if err != nil {
		s.logger.Info("no session-id found in cookie")
	}

	return v.Value
}
