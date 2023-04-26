package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/lmittmann/tint"
	"golang.org/x/exp/slog"
)

var (
  addr = "127.0.0.1:8080"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout))

  if val, ok := os.LookupEnv("ADDR"); ok {
    addr = val
  }

	router := Router{
		Logger: logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", router.handleStatic())
	mux.HandleFunc("/", router.handleRoot)

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	idleConnClosed := make(chan any)
	go gracefulShutdown(&server, logger, idleConnClosed)

	logger.Info("server is listening", "addr", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server unexpectedly quit", "err", err)
		os.Exit(1)
	}

	<-idleConnClosed
	logger.Info("server terminated succesfully")
}

// gracefulShutdown shutsdown the server after receiving a interrupt signal
// and closes the closed channel to signal succesful termination af idle connections
// if timeout is reached, idle connections are closed forcefully
func gracefulShutdown(server *http.Server, logger *slog.Logger, closed chan any) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("gracefully closing idle connections")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to close idle connections", "err", err)
		os.Exit(1)
	}

	close(closed)
}
