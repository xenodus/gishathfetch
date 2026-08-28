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

// SetMyCommands registers slash commands shown in the Telegram command menu.
func (t *TelegramAPI) SetMyCommands(ctx context.Context, commands []BotCommand) error {
	payload := map[string]any{
		"commands": commands,
	}
	return t.post(ctx, "setMyCommands", payload)
}

// SetWebhook registers the bot webhook URL and secret token with Telegram.
func (t *TelegramAPI) SetWebhook(ctx context.Context, webhookURL, secretToken string) error {
	payload := map[string]string{
		"url":          webhookURL,
		"secret_token": secretToken,
	}
	return t.post(ctx, "setWebhook", payload)
}

// SendForceReply asks the user to reply with text. Telegram opens the input
// field focused on this chat so the user can type their answer.
func (t *TelegramAPI) SendForceReply(ctx context.Context, chatID int64, text, placeholder, parseMode string) (int64, error) {
	payload := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"disable_web_page_preview": true,
		"reply_markup": map[string]any{
			"force_reply":             true,
			"input_field_placeholder": strings.TrimSpace(placeholder),
		},
	}
	if mode := strings.TrimSpace(parseMode); mode != "" {
		payload["parse_mode"] = mode
	}
	return t.sendMessage(ctx, payload)
}

// SendMessage sends a text message to a chat.
// When linkPreviewURL is non-empty, Telegram previews that URL (it must appear
// in text). Otherwise Telegram uses the first URL found in the message.
func (t *TelegramAPI) SendMessage(ctx context.Context, chatID int64, text, linkPreviewURL string) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	if preview := strings.TrimSpace(linkPreviewURL); preview != "" {
		payload["link_preview_options"] = map[string]any{
			"url": preview,
		}
	} else {
		payload["disable_web_page_preview"] = true
	}
	_, err := t.sendMessage(ctx, payload)
	return err
}

// SendPhoto sends a photo to a chat. photoURL must be a Telegram-supported file_id or HTTP URL.
func (t *TelegramAPI) SendPhoto(ctx context.Context, chatID int64, photoURL, caption string) error {
	payload := map[string]any{
		"chat_id": chatID,
		"photo":   strings.TrimSpace(photoURL),
	}
	if caption = strings.TrimSpace(caption); caption != "" {
		payload["caption"] = caption
	}
	return t.post(ctx, "sendPhoto", payload)
}

func (t *TelegramAPI) sendMessage(ctx context.Context, payload map[string]any) (int64, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := t.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("telegram sendMessage failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var envelope struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return 0, err
	}
	if !envelope.OK {
		if envelope.Description == "" {
			envelope.Description = "telegram api error"
		}
		return 0, fmt.Errorf("telegram sendMessage failed: %s", envelope.Description)
	}
	return envelope.Result.MessageID, nil
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
