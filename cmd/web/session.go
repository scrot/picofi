package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/allegro/bigcache/v3"
	"github.com/google/uuid"
)

var (
	ErrSessionExpired = errors.New("session expired and data purged from cache")
)

// Session is the data that can be customized by the user to personalize
// the calculations, it is used to populate the templates.
type Session struct {
	AnnualIncome   float64      `json:"annual-income"`
	AnnualExpenses float64      `json:"annual-expenses"`
	Savings        []SavingsRow `json:"savings"`
}

// toJSON serializes the SimulationInput as JSON object
func (s Session) toJSON() ([]byte, error) {
	r, err := json.Marshal(s)
	if err != nil {
		return []byte{}, nil
	}
	return r, nil
}

// fromContext loads session from context if present
func fromContext(ctx context.Context) (Session, error) {
	v := ctx.Value(sessionContexKey)

	session, ok := v.(Session)
	if !ok {
		return Session{}, fmt.Errorf("unable to cast to Session")
	}

	return session, nil
}

// SavingsRow represents a money deposit with optional interests
type SavingsRow struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Interest float64 `json:"interest"`
}

// sessionDefaults provides default values for a session
// used for setting up new session and contain sane values
var sessionDefaults = Session{
	AnnualIncome:   70000,
	AnnualExpenses: 50000,
	Savings: []SavingsRow{
		{"Savings Account", 15000, 2},
	},
}

// sessionCookieKey is the default cookie key for storing session
var sessionCookieKey = "session-id"

// sessionContextKey use as key to store Session data in context
type ContextKey string

var sessionContexKey = ContextKey("session-id")

// newSession generates a new UUID session id and load session defaults into cache
func (s Server) newSession() string {
	id := uuid.New()
	defaults, _ := sessionDefaults.toJSON()
	s.sessions.Set(id.String(), defaults)
	return id.String()
}

// loadSession loads the session data given a id
// return ErrSessionExpired if session is not found in cache.
func (s Server) loadSession(id string) (Session, error) {
	s.logger.Info(fmt.Sprintf("retreiving session data of %s", id))

	var data Session
	sd, err := s.sessions.Get(id)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			s.logger.Info(fmt.Sprintf("session %s not found in cache", id))
			return Session{}, ErrSessionExpired
		} else {
			return Session{}, fmt.Errorf("unexpected error loading session data: %w", err)
		}
	}

	if err := json.Unmarshal(sd, &data); err != nil {
		return Session{}, fmt.Errorf("parsing session data: %w", err)
	}

	return data, nil
}

// saveSession stores update in the cache, overwriting the old entry
func (s Server) saveSession(id string, update Session) error {
	s.logger.Info(fmt.Sprintf("updating session data of %s", id))

	obj, err := update.toJSON()
	if err != nil {
		return fmt.Errorf("serializing updated session data: %w", err)
	}

	if err := s.sessions.Set(id, obj); err != nil {
		return fmt.Errorf("storing updated session data: %w", err)
	}

	return nil
}
