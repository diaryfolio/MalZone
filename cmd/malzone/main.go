package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diaryfolio/malzone/internal/api"
	"github.com/diaryfolio/malzone/internal/kube"
	"github.com/diaryfolio/malzone/internal/operator"
	"github.com/diaryfolio/malzone/internal/runner"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if len(os.Args) != 2 {
		logger.Error("usage: malzone api|operator|runner")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	var err error
	switch os.Args[1] {
	case "api":
		err = runAPI(ctx, logger)
	case "operator":
		err = runOperator(ctx, logger)
	case "runner":
		err = runner.Run(ctx)
	default:
		logger.Error("unknown mode", "mode", os.Args[1])
		os.Exit(2)
	}
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("component stopped", "mode", os.Args[1], "error", err)
		os.Exit(1)
	}
}

func runAPI(ctx context.Context, logger *slog.Logger) error {
	namespace := envOr("POD_NAMESPACE", "malzone-system")
	client, err := kube.NewInCluster(namespace)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:              ":8080",
		Handler:           api.New(client, namespace, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	logger.Info("API starting", "address", server.Addr, "namespace", namespace, "contract", "v1alpha1-poc")
	return server.ListenAndServe()
}

func runOperator(ctx context.Context, logger *slog.Logger) error {
	namespace := envOr("POD_NAMESPACE", "malzone-system")
	client, err := kube.NewInCluster(namespace)
	if err != nil {
		return err
	}
	runnerImage := envOr("RUNNER_IMAGE", "malzone-poc:dev")
	logger.Info("operator starting", "namespace", namespace, "backend", "kubernetes-job-poc")
	return operator.New(client, runnerImage, time.Second, logger).Run(ctx)
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
