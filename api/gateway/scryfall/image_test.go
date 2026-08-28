package scryfall

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLookupImageURL(t *testing.T) {
	originalHTTPGet := httpGet
	t.Cleanup(func() {
		httpGet = originalHTTPGet
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/cards/named", r.URL.Path)
		require.Equal(t, "Lightning Bolt", r.URL.Query().Get("fuzzy"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name": "Lightning Bolt",
			"image_uris": map[string]string{
				"normal": "https://cards.scryfall.io/normal/front/a/b/c.jpg",
			},
		})
	}))
	t.Cleanup(server.Close)

	httpGet = func(ctx context.Context, requestURL string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/cards/named?fuzzy=Lightning+Bolt", nil)
		require.NoError(t, err)
		return http.DefaultClient.Do(req)
	}

	imageURL, err := LookupImageURL(context.Background(), "Lightning Bolt")
	require.NoError(t, err)
	require.Equal(t, "https://cards.scryfall.io/normal/front/a/b/c.jpg", imageURL)
}

func TestLookupImageURL_DoubleFaced(t *testing.T) {
	originalHTTPGet := httpGet
	t.Cleanup(func() {
		httpGet = originalHTTPGet
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name": "Delver of Secrets",
			"card_faces": []map[string]any{
				{
					"image_uris": map[string]string{
						"normal": "https://cards.scryfall.io/normal/front/d/e/f.jpg",
					},
				},
			},
		})
	}))
	t.Cleanup(server.Close)

	httpGet = func(ctx context.Context, requestURL string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/cards/named?fuzzy=Delver+of+Secrets", nil)
		require.NoError(t, err)
		return http.DefaultClient.Do(req)
	}

	imageURL, err := LookupImageURL(context.Background(), "Delver of Secrets")
	require.NoError(t, err)
	require.Equal(t, "https://cards.scryfall.io/normal/front/d/e/f.jpg", imageURL)
}

func TestLookupImageURL_EmptyName(t *testing.T) {
	imageURL, err := LookupImageURL(context.Background(), "  ")
	require.NoError(t, err)
	require.Empty(t, imageURL)
}
