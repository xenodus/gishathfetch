package telegrambot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TelegramAPI is a minimal Telegram Bot API client.
type TelegramAPI struct {
	token      string
	httpClient *http.Client
}

// NewTelegramAPI creates a client for api.telegram.org.
func NewTelegramAPI(token string) *TelegramAPI {
	return &TelegramAPI{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetWebhook registers the bot webhook URL and secret token with Telegram.
func (t *TelegramAPI) SetWebhook(ctx context.Context, webhookURL, secretToken string) error {
	payload := map[string]string{
		"url":          webhookURL,
		"secret_token": secretToken,
	}
	return t.post(ctx, "setWebhook", payload)
}

// SendMessage sends a text message to a chat.
func (t *TelegramAPI) SendMessage(ctx context.Context, chatID int64, text string) error {
	payload := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"disable_web_page_preview": false,
	}
	return t.post(ctx, "sendMessage", payload)
}

func (t *TelegramAPI) post(ctx context.Context, method string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram %s failed: status %d: %s", method, res.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var envelope struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return err
	}
	if !envelope.OK {
		if envelope.Description == "" {
			envelope.Description = "telegram api error"
		}
		return fmt.Errorf("telegram %s failed: %s", method, envelope.Description)
	}
	return nil
}
