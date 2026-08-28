package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"mtg-price-checker-sg/pkg/telegrambot"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := telegrambot.LoadConfig()
	if cfg.TelegramBotToken == "" {
		logger.Error("TELEGRAM_BOT_TOKEN is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	api := telegrambot.NewTelegramAPI(cfg.TelegramBotToken)
	if err := telegrambot.RegisterBot(ctx, cfg, api); err != nil {
		logger.Error("telegram bot registration failed", "err", err)
		os.Exit(1)
	}

	logger.Info("telegram bot commands registered", "count", len(telegrambot.TelegramMenuCommands()))
	if cfg.WebhookPublicURL != "" && cfg.WebhookSecret != "" {
		logger.Info("telegram webhook registered", "url", cfg.WebhookPublicURL)
	}
}
