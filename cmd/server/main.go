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
	addr = os.Getenv("ADDR")
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout))

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)

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

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}
