package telegrambot

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_HandleWebhook_SyncPrice(t *testing.T) {
	var searched string
	gishath := &stubGishath{
		search: func(_ context.Context, query string) (*SearchSummary, error) {
			searched = query
			return &SearchSummary{
				ResultCount: 1,
				Cheapest: &CardSummary{
					Name:   "Opt",
					Price:  1.0,
					Source: "Flagship Games",
				},
				WebsiteURL: "https://gishathfetch.com/?s=Opt",
			}, nil
		},
	}
	var messages []string
	telegram := &stubTelegram{
		send: func(_ context.Context, _ int64, text string) error {
			messages = append(messages, text)
			return nil
		},
	}

	svc := NewService("webhook-secret", gishath, telegram, nil, slog.Default())
	body, err := json.Marshal(Update{
		Message: &Message{
			Chat: Chat{ID: 42},
			Text: "/price Opt",
		},
	})
	require.NoError(t, err)

	status := svc.HandleWebhook(context.Background(), "webhook-secret", body)
	require.Equal(t, WebhookStatusOK, status)
	require.Equal(t, "Opt", searched)
	require.GreaterOrEqual(t, len(messages), 2)
	require.Contains(t, messages[0], "Searching")
	require.Contains(t, messages[len(messages)-1], "S$1.00")
}

func TestService_HandleWebhook_AsyncPrice(t *testing.T) {
	var enqueued bool
	async := &stubAsync{
		enqueue: func(_ context.Context, chatID int64, query string) error {
			require.Equal(t, int64(42), chatID)
			require.Equal(t, "Opt", query)
			enqueued = true
			return nil
		},
	}
	var messages []string
	telegram := &stubTelegram{
		send: func(_ context.Context, _ int64, text string) error {
			messages = append(messages, text)
			return nil
		},
	}

	svc := NewService("webhook-secret", &stubGishath{}, telegram, async, slog.Default())
	body, err := json.Marshal(Update{
		Message: &Message{
			Chat: Chat{ID: 42},
			Text: "/price Opt",
		},
	})
	require.NoError(t, err)

	status := svc.HandleWebhook(context.Background(), "webhook-secret", body)
	require.Equal(t, WebhookStatusOK, status)
	require.True(t, enqueued)
	require.Len(t, messages, 1)
	require.Contains(t, messages[0], "Searching")
}

func TestHandler_ForbiddenWithoutSecret(t *testing.T) {
	svc := NewService("webhook-secret", &stubGishath{}, &stubTelegram{}, nil, slog.Default())
	handler := NewHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", bytes.NewReader([]byte("{}")))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestService_RunPriceSearch(t *testing.T) {
	gishath := &stubGishath{
		search: func(_ context.Context, _ string) (*SearchSummary, error) {
			return &SearchSummary{
				ResultCount: 0,
				WebsiteURL:  "https://gishathfetch.com/?s=zzz",
			}, nil
		},
	}
	var sent string
	telegram := &stubTelegram{
		send: func(_ context.Context, _ int64, text string) error {
			sent = text
			return nil
		},
	}

	svc := NewService("secret", gishath, telegram, nil, slog.Default())
	require.NoError(t, svc.RunPriceSearch(context.Background(), 1, "zzz"))
	require.Contains(t, sent, "No in-stock matches")
	require.NotContains(t, sent, "View on Gishath Fetch")
	require.NotContains(t, sent, "gishathfetch.com")
}

func Test_formatSearchReply(t *testing.T) {
	reply := formatSearchReply("Opt", &SearchSummary{
		ResultCount: 2,
		Cheapest: &CardSummary{
			Name:    "Opt",
			Price:   1.25,
			Source:  "Hideout",
			Quality: "NM",
			URL:     "https://shop.example/opt",
		},
		WebsiteURL: "https://gishathfetch.com/?s=Opt",
	})
	require.Contains(t, reply, "S$1.25")
	require.Contains(t, reply, "2 results")
	require.Contains(t, reply, "https://gishathfetch.com/?s=Opt")
}

func Test_formatSearchReply_NoMatches(t *testing.T) {
	reply := formatSearchReply("zzz", &SearchSummary{
		ResultCount: 0,
		WebsiteURL:  "https://gishathfetch.com/?s=zzz",
	})
	require.Equal(t, `No in-stock matches for "zzz".`, reply)
	require.NotContains(t, reply, "View on Gishath Fetch")
	require.NotContains(t, reply, "gishathfetch.com")
}

type stubGishath struct {
	search func(context.Context, string) (*SearchSummary, error)
}

func (s *stubGishath) Search(ctx context.Context, query string) (*SearchSummary, error) {
	if s.search != nil {
		return s.search(ctx, query)
	}
	return nil, nil
}

type stubTelegram struct {
	send func(context.Context, int64, string) error
}

func (s *stubTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	if s.send != nil {
		return s.send(ctx, chatID, text)
	}
	return nil
}

type stubAsync struct {
	enqueue func(context.Context, int64, string) error
}

func (s *stubAsync) EnqueuePriceSearch(ctx context.Context, chatID int64, query string) error {
	if s.enqueue != nil {
		return s.enqueue(ctx, chatID, query)
	}
	return nil
}
