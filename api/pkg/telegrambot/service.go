package telegrambot

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"mtg-price-checker-sg/pkg/telegramphoto"
)

const telegramSecretHeader = "X-Telegram-Bot-Api-Secret-Token"

// SecretHeader is the Telegram webhook secret header name.
const SecretHeader = telegramSecretHeader

// Service processes Telegram webhook updates.
type Service struct {
	secret   string
	gishath  searchClient
	telegram messageClient
	async    PriceSearchAsync
	logger   *slog.Logger
}

type searchClient interface {
	Search(ctx context.Context, query string) (*SearchSummary, error)
}

type messageClient interface {
	SendMessage(ctx context.Context, chatID int64, text, linkPreviewURL string) error
	SendPhoto(ctx context.Context, chatID int64, photoURL, caption string) error
}

// NewService wires Telegram webhook handling. When async is nil, /price runs synchronously.
func NewService(
	secret string,
	gishath searchClient,
	telegram messageClient,
	async PriceSearchAsync,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		secret:   secret,
		gishath:  gishath,
		telegram: telegram,
		async:    async,
		logger:   logger,
	}
}

// HandleWebhook validates the Telegram secret token and processes one update.
// It returns an HTTP status code suitable for the webhook response.
func (s *Service) HandleWebhook(ctx context.Context, secretHeader string, body []byte) int {
	if !s.validSecret(secretHeader) {
		return WebhookStatusForbidden
	}

	var update Update
	if err := json.Unmarshal(body, &update); err != nil {
		return WebhookStatusBadRequest
	}

	if update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
		return WebhookStatusOK
	}

	status, err := s.handleMessage(ctx, update.Message)
	if err != nil {
		s.logger.ErrorContext(ctx, "telegram command failed", "err", err)
		return WebhookStatusInternalError
	}
	return status
}

func (s *Service) validSecret(got string) bool {
	got = strings.TrimSpace(got)
	expected := strings.TrimSpace(s.secret)
	if got == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}

func (s *Service) handleMessage(ctx context.Context, message *Message) (int, error) {
	text := strings.TrimSpace(message.Text)
	if text == "" {
		return WebhookStatusOK, nil
	}

	switch {
	case strings.EqualFold(text, "/help"):
		if err := s.telegram.SendMessage(ctx, message.Chat.ID, formatHelpMessage(), ""); err != nil {
			return WebhookStatusInternalError, err
		}
		return WebhookStatusOK, nil
	case strings.HasPrefix(text, "/price"):
		query := strings.TrimSpace(strings.TrimPrefix(text, "/price"))
		if query == "" {
			if err := s.telegram.SendMessage(ctx, message.Chat.ID, "Usage: /price <card name>", ""); err != nil {
				return WebhookStatusInternalError, err
			}
			return WebhookStatusOK, nil
		}
		if len(query) < 3 {
			if err := s.telegram.SendMessage(ctx, message.Chat.ID, "Enter at least 3 characters to search.", ""); err != nil {
				return WebhookStatusInternalError, err
			}
			return WebhookStatusOK, nil
		}

		if err := s.telegram.SendMessage(ctx, message.Chat.ID, "Searching for "+query+"…", ""); err != nil {
			return WebhookStatusInternalError, err
		}

		if s.async != nil {
			if err := s.async.EnqueuePriceSearch(ctx, message.Chat.ID, query); err != nil {
				return WebhookStatusInternalError, err
			}
			return WebhookStatusOK, nil
		}

		if err := s.RunPriceSearch(ctx, message.Chat.ID, query); err != nil {
			return WebhookStatusInternalError, err
		}
		return WebhookStatusOK, nil
	default:
		return WebhookStatusOK, nil
	}
}

// RunPriceSearch performs the Gishath lookup and sends the Telegram reply.
func (s *Service) RunPriceSearch(ctx context.Context, chatID int64, query string) error {
	summary, err := s.gishath.Search(ctx, query)
	if err != nil {
		s.logger.ErrorContext(ctx, "telegram price search failed", "query", query, "err", err)
		return s.telegram.SendMessage(ctx, chatID, "Search failed. Please try again later.", "")
	}
	caption := formatSearchReply(query, summary)
	return s.sendPriceSearchReply(ctx, chatID, summary, caption)
}

func (s *Service) sendPriceSearchReply(ctx context.Context, chatID int64, summary *SearchSummary, caption string) error {
	if summary == nil || summary.ResultCount == 0 || summary.Cheapest == nil {
		return s.telegram.SendMessage(ctx, chatID, caption, "")
	}

	previewURL := strings.TrimSpace(summary.WebsiteURL)
	photoURL := strings.TrimSpace(summary.PhotoURL)
	if photoURL == "" && summary.Cheapest != nil {
		photoURL = telegramphoto.Normalize(summary.Cheapest.Img)
	}
	if telegramphoto.IsSendable(photoURL) {
		if err := s.telegram.SendPhoto(ctx, chatID, photoURL, caption); err != nil {
			s.logger.WarnContext(ctx, "telegram sendPhoto failed, falling back to sendMessage",
				"photoURL", photoURL, "err", err)
			return s.telegram.SendMessage(ctx, chatID, caption, previewURL)
		}
		return nil
	}
	return s.telegram.SendMessage(ctx, chatID, caption, previewURL)
}

// Handler adapts Service to net/http for local development servers.
type Handler struct {
	service *Service
}

// NewHandler returns an http.Handler backed by Service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ServeHTTP validates the webhook secret and handles supported commands.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := readWebhookBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	status := h.service.HandleWebhook(r.Context(), r.Header.Get(telegramSecretHeader), body)
	if status == WebhookStatusOK {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, http.StatusText(status), status)
}

func readWebhookBody(r *http.Request) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r.Body, 1<<20))
}
