package telegrambot

import (
	"os"
	"strings"
)

// Config holds runtime settings for the Telegram webhook service.
type Config struct {
	TelegramBotToken     string
	WebhookSecret        string
	WebhookPath          string
	WebhookPublicURL     string
	ListenAddr           string
	GishathAPIBaseURL    string
	GishathBotToken      string
	GishathOriginSecret  string
}

const (
	defaultAPIBaseURL   = "https://api.gishathfetch.com"
	defaultListenAddr   = ":8080"
	defaultWebhookPath  = "/telegram/webhook"
)

// LoadConfig reads Telegram bot settings from the environment.
func LoadConfig() Config {
	webhookPath := strings.TrimSpace(os.Getenv("TELEGRAM_WEBHOOK_PATH"))
	if webhookPath == "" {
		webhookPath = defaultWebhookPath
	}
	if !strings.HasPrefix(webhookPath, "/") {
		webhookPath = "/" + webhookPath
	}

	apiBase := strings.TrimSpace(os.Getenv("GISHATH_API_BASE_URL"))
	if apiBase == "" {
		apiBase = defaultAPIBaseURL
	}
	apiBase = strings.TrimRight(apiBase, "/")

	listenAddr := strings.TrimSpace(os.Getenv("TELEGRAM_LISTEN_ADDR"))
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	return Config{
		TelegramBotToken:    strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		WebhookSecret:       strings.TrimSpace(os.Getenv("TELEGRAM_WEBHOOK_SECRET")),
		WebhookPath:         webhookPath,
		WebhookPublicURL:    strings.TrimSpace(os.Getenv("TELEGRAM_WEBHOOK_PUBLIC_URL")),
		ListenAddr:          listenAddr,
		GishathAPIBaseURL:   apiBase,
		GishathBotToken:     strings.TrimSpace(os.Getenv("API_TELEGRAM_BOT_TOKEN")),
		GishathOriginSecret: strings.TrimSpace(os.Getenv("API_ORIGIN_VERIFY_SECRET")),
	}
}

func (c Config) Valid() error {
	if c.TelegramBotToken == "" {
		return errConfig("TELEGRAM_BOT_TOKEN is required")
	}
	if c.WebhookSecret == "" {
		return errConfig("TELEGRAM_WEBHOOK_SECRET is required")
	}
	if c.GishathBotToken == "" {
		return errConfig("API_TELEGRAM_BOT_TOKEN is required")
	}
	return nil
}

type configError string

func errConfig(message string) error {
	return configError(message)
}

func (e configError) Error() string {
	return string(e)
}
