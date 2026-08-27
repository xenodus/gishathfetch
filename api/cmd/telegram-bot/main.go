package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mtg-price-checker-sg/pkg/telegrambot"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := telegrambot.LoadConfig()
	svc, err := telegrambot.NewServiceFromConfig(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	telegram := telegrambot.NewTelegramAPI(cfg.TelegramBotToken)
	if cfg.WebhookPublicURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := telegram.SetWebhook(ctx, cfg.WebhookPublicURL, cfg.WebhookSecret); err != nil {
			logger.Error("failed to set telegram webhook", "err", err)
			os.Exit(1)
		}
		logger.Info("telegram webhook registered", "url", cfg.WebhookPublicURL)
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.WebhookPath, telegrambot.NewHandler(svc))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("telegram bot listening", "addr", cfg.ListenAddr, "webhook_path", cfg.WebhookPath)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", "err", err)
		os.Exit(1)
	}
}
