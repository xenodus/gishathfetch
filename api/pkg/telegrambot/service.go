package telegrambot

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"mtg-price-checker-sg/pkg/telegramphoto"
)

const (
	telegramSecretHeader   = "X-Telegram-Bot-Api-Secret-Token"
	pricePromptMessage     = "Enter a card name to search:"
	pricePromptPlaceholder = "Lightning Bolt"
)


// SecretHeader is the Telegram webhook secret header name.
const SecretHeader = telegramSecretHeader

// Service processes Telegram webhook updates.
type Service struct {
	secret   string
	gishath  searchClient
	telegram messageClient
	async    PriceSearchAsync
	logger   *slog.Logger

	pendingPriceChats sync.Map // pendingPriceKey -> pendingPricePrompt
}

type pendingPriceKey struct {
	chatID int64
	userID int64
}

type pendingPricePrompt struct {
	messageID int64
}

type searchClient interface {
	Search(ctx context.Context, query string) (*SearchSummary, error)
}

type messageClient interface {
	SendMessage(ctx context.Context, chatID int64, text, linkPreviewURL string) error
	SendPhoto(ctx context.Context, chatID int64, photoURL, caption string) error
	SendForceReply(ctx context.Context, chatID int64, text, placeholder string) (int64, error)
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

	chatID := message.Chat.ID
	userID := message.SenderID()

	switch {
	case strings.EqualFold(text, "/help"):
		s.clearPendingPrice(chatID, userID)
		if err := s.telegram.SendMessage(ctx, chatID, formatHelpMessage(), ""); err != nil {
			return WebhookStatusInternalError, err
		}
		return WebhookStatusOK, nil
	case strings.HasPrefix(text, "/price"):
		query := strings.TrimSpace(strings.TrimPrefix(text, "/price"))
		if query == "" {
			return s.promptPriceQuery(ctx, chatID, userID)
		}
		s.clearPendingPrice(chatID, userID)
		return s.handlePriceQuery(ctx, chatID, userID, query)
	}

	if s.isFollowUpToPricePrompt(message) {
		s.clearPendingPrice(chatID, userID)
		return s.handlePriceQuery(ctx, chatID, userID, text)
	}

	return WebhookStatusOK, nil
}

func (s *Service) promptPriceQuery(ctx context.Context, chatID, userID int64) (int, error) {
	messageID, err := s.telegram.SendForceReply(ctx, chatID, pricePromptMessage, pricePromptPlaceholder)
	if err != nil {
		return WebhookStatusInternalError, err
	}
	s.setPendingPrice(chatID, userID, messageID)
	return WebhookStatusOK, nil
}

func (s *Service) pendingPriceKey(chatID, userID int64) pendingPriceKey {
	return pendingPriceKey{chatID: chatID, userID: userID}
}

func (s *Service) setPendingPrice(chatID, userID, messageID int64) {
	s.pendingPriceChats.Store(s.pendingPriceKey(chatID, userID), pendingPricePrompt{messageID: messageID})
}

func (s *Service) clearPendingPrice(chatID, userID int64) {
	s.pendingPriceChats.Delete(s.pendingPriceKey(chatID, userID))
}

func (s *Service) hasPendingPrice(chatID, userID int64) bool {
	_, ok := s.pendingPriceChats.Load(s.pendingPriceKey(chatID, userID))
	return ok
}

func (s *Service) isFollowUpToPricePrompt(message *Message) bool {
	chatID := message.Chat.ID
	userID := message.SenderID()
	if !s.hasPendingPrice(chatID, userID) {
		return false
	}
	if message.ReplyToMessage != nil {
		return s.replyTargetsPricePrompt(message)
	}
	return true
}

func (s *Service) replyTargetsPricePrompt(message *Message) bool {
	reply := message.ReplyToMessage
	if reply == nil {
		return false
	}
	if isPricePrompt(reply.Text) {
		return true
	}
	if reply.MessageID == 0 {
		return false
	}
	pending, ok := s.pendingPriceChats.Load(s.pendingPriceKey(message.Chat.ID, message.SenderID()))
	if !ok {
		return false
	}
	return pending.(pendingPricePrompt).messageID == reply.MessageID
}

func isPricePrompt(text string) bool {
	return strings.TrimSpace(text) == pricePromptMessage
}

func (s *Service) handlePriceQuery(ctx context.Context, chatID, userID int64, query string) (int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.promptPriceQuery(ctx, chatID, userID)
	}
	if len(query) < 3 {
		s.setPendingPrice(chatID, userID, 0)
		if err := s.telegram.SendMessage(ctx, chatID, "Enter at least 3 characters to search.", ""); err != nil {
			return WebhookStatusInternalError, err
		}
		return WebhookStatusOK, nil
	}

	if err := s.telegram.SendMessage(ctx, chatID, "Searching for "+query+"…", ""); err != nil {
		return WebhookStatusInternalError, err
	}

	if s.async != nil {
		if err := s.async.EnqueuePriceSearch(ctx, chatID, query); err != nil {
			return WebhookStatusInternalError, err
		}
		return WebhookStatusOK, nil
	}

	if err := s.RunPriceSearch(ctx, chatID, query); err != nil {
		return WebhookStatusInternalError, err
	}
	return WebhookStatusOK, nil
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

	photoURL := strings.TrimSpace(summary.PhotoURL)
	if photoURL == "" && summary.Cheapest != nil {
		photoURL = telegramphoto.Normalize(summary.Cheapest.Img)
	}
	if telegramphoto.IsSendable(photoURL) {
		if err := s.telegram.SendPhoto(ctx, chatID, photoURL, caption); err != nil {
			s.logger.WarnContext(ctx, "telegram sendPhoto failed, falling back to sendMessage",
				"photoURL", photoURL, "err", err)
			return s.telegram.SendMessage(ctx, chatID, caption, "")
		}
		return nil
	}
	return s.telegram.SendMessage(ctx, chatID, caption, "")
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
