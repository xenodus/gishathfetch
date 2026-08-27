package telegrambot

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const telegramSecretHeader = "X-Telegram-Bot-Api-Secret-Token"

// Update is a subset of the Telegram webhook update payload.
type Update struct {
	Message *Message `json:"message"`
}

// Message is a subset of a Telegram message.
type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

// Chat identifies a Telegram chat.
type Chat struct {
	ID int64 `json:"id"`
}

// Handler serves the Telegram webhook and processes bot commands.
type Handler struct {
	secret      string
	gishath     searchClient
	telegram    messageClient
	logger      *slog.Logger
}

type searchClient interface {
	Search(ctx context.Context, query string) (*SearchSummary, error)
}

type messageClient interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

// NewHandler wires command handling for Telegram updates.
func NewHandler(secret string, gishath searchClient, telegram messageClient, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		secret:   secret,
		gishath:  gishath,
		telegram: telegram,
		logger:   logger,
	}
}

// ServeHTTP validates the webhook secret and handles supported commands.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.validSecret(r.Header.Get(telegramSecretHeader)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var update Update
	if err := json.Unmarshal(body, &update); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if update.Message == nil || update.Message.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.handleMessage(r.Context(), update.Message); err != nil {
		h.logger.ErrorContext(r.Context(), "telegram command failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) validSecret(got string) bool {
	got = strings.TrimSpace(got)
	expected := strings.TrimSpace(h.secret)
	if got == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}

func (h *Handler) handleMessage(ctx context.Context, message *Message) error {
	text := strings.TrimSpace(message.Text)
	if text == "" {
		return nil
	}

	switch {
	case strings.EqualFold(text, "/help"):
		return h.telegram.SendMessage(ctx, message.Chat.ID, formatHelpMessage())
	case strings.HasPrefix(text, "/price"):
		query := strings.TrimSpace(strings.TrimPrefix(text, "/price"))
		if query == "" {
			return h.telegram.SendMessage(ctx, message.Chat.ID, "Usage: /price <card name>")
		}
		if len(query) < 3 {
			return h.telegram.SendMessage(ctx, message.Chat.ID, "Enter at least 3 characters to search.")
		}

		summary, err := h.gishath.Search(ctx, query)
		if err != nil {
			return err
		}
		return h.telegram.SendMessage(ctx, message.Chat.ID, formatSearchReply(query, summary))
	default:
		return nil
	}
}
