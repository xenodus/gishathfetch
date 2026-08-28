package telegrambot

import (
	"context"
	"strings"
)

// RegisterBot updates Telegram bot metadata: slash commands and, when configured,
// the webhook URL and secret token.
func RegisterBot(ctx context.Context, cfg Config, api *TelegramAPI) error {
	if err := api.SetMyCommands(ctx, TelegramMenuCommands()); err != nil {
		return err
	}

	webhookURL := strings.TrimSpace(cfg.WebhookPublicURL)
	webhookSecret := strings.TrimSpace(cfg.WebhookSecret)
	if webhookURL != "" && webhookSecret != "" {
		return api.SetWebhook(ctx, webhookURL, webhookSecret)
	}
	return nil
}
