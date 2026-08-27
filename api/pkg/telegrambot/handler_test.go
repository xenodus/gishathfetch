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

func TestHandler_WebhookSecret(t *testing.T) {
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
	var sent string
	telegram := &stubTelegram{
		send: func(_ context.Context, _ int64, text string) error {
			sent = text
			return nil
		},
	}

	handler := NewHandler("webhook-secret", gishath, telegram, slog.Default())
	body, err := json.Marshal(Update{
		Message: &Message{
			Chat: Chat{ID: 42},
			Text: "/price Opt",
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", bytes.NewReader(body))
	req.Header.Set(telegramSecretHeader, "webhook-secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "Opt", searched)
	require.Contains(t, sent, "S$1.00")
}

func TestHandler_ForbiddenWithoutSecret(t *testing.T) {
	handler := NewHandler("webhook-secret", &stubGishath{}, &stubTelegram{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", bytes.NewReader([]byte("{}")))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
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
