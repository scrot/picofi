package main

import (
	"context"
	"errors"
	"net/http"
)

func (s Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookieKey)

		// load cookie, if none found generate new session
		var id string
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				id = s.generateNewSession(w)
			} else {
				s.logger.Error("inspecting session cookie", "err", err)
				return
			}
		} else {
			id = c.Value
		}

		// load data, if expired generate new session
		session, err := s.loadSession(id)
		if err != nil {
			if errors.Is(err, ErrSessionExpired) {
				id = s.generateNewSession(w)
			} else {
				s.logger.Error("loading session", "err", err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), sessionContexKey, session)
		r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (s Server) generateNewSession(w http.ResponseWriter) string {
	id := s.newSession()

	c := &http.Cookie{Name: sessionCookieKey, Value: id}
	http.SetCookie(w, c)

	s.logger.Info("new session created", "session", id)

	return id
}
