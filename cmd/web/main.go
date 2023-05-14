package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/lmittmann/tint"
	"github.com/scrot/picofi"
	"golang.org/x/exp/slog"
)

var (
	addr = "127.0.0.1:8080"
)

func main() {
	if val, ok := os.LookupEnv("ADDR"); ok {
		addr = val
	}

	logger := slog.New(tint.NewHandler(os.Stdout))

	money.AddCurrency("EURA", "\u20ac", "$1", ",", ".", 0)
	calculator := picofi.NewCalculator(logger, *money.GetCurrency("EURA"))

	router := NewServer(logger, calculator)

	mux := http.NewServeMux()
	mux.Handle("/static/", router.handleStatic())

	mux.HandleFunc("/", router.handleOverview)
	mux.HandleFunc("/past", router.handlePast)
	mux.Handle("/simulation", router.sessionMiddleware(http.HandlerFunc(router.handleSimulation)))
	mux.HandleFunc("/simulation/savings", router.updateSavings)

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	idleConnClosed := make(chan any)
	go gracefulShutdown(&server, logger, idleConnClosed)

	logger.Info("main: server is listening", "addr", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("main: server unexpectedly quit", "err", err)
		os.Exit(1)
	}

	<-idleConnClosed
	logger.Info("main: server terminated succesfully")
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

	logger.Info("gracefulShutdown: gracefully closing idle connections")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("gracefulShutdown: failed to close idle connections", "err", err)
		os.Exit(1)
	}

	close(closed)
}
