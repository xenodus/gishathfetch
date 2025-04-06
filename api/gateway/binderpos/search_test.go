package binderpos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Search_Success(t *testing.T) {
	givenStoreName := "MTG Asia"
	givenStoreBaseURL := "https://www.mtg-asia.com"
	givePayload, err := json.Marshal(Payload{
		StoreURL:    "mtgasia.myshopify.com",
		Game:        ProductTypeMTG.ToString(),
		Title:       "Abrade",
		InstockOnly: true,
	})
	require.NoError(t, err)

	i := New()
	cards, httpStatusCode, err := i.Search(givenStoreName, givenStoreBaseURL, givePayload)

	require.NoError(t, err)
	require.True(t, len(cards) > 0)
	require.Equal(t, http.StatusOK, httpStatusCode)

	for _, card := range cards {
		if card.InStock {
			require.NotEmpty(t, card.Name)
			require.NotEmpty(t, card.Source)
			require.NotEmpty(t, card.Url)
			require.NotEmpty(t, card.Img)
			require.NotEmpty(t, card.Price)
			require.Contains(t, card.Url, givenStoreBaseURL+productPath)
		}
	}
}

func Test_Search_HttpFailure(t *testing.T) {
	// Mock the Binderpos API to return a 400 error
	mockBinderposSearch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{}`))
	}))
	defer mockBinderposSearch.Close()

	givenStoreName := "ASDF"
	givenStoreBaseURL := "ASDF"
	givePayload, err := json.Marshal(Payload{
		StoreURL:    "ASDF",
		Game:        ProductTypeMTG.ToString(),
		Title:       "Abrade",
		InstockOnly: true,
	})
	require.NoError(t, err)

	i := NewWithApiUrl(mockBinderposSearch.URL)
	cards, httpStatusCode, err := i.Search(givenStoreName, givenStoreBaseURL, givePayload)

	require.NoError(t, err)
	require.Equal(t, 0, len(cards))
	require.Equal(t, http.StatusBadRequest, httpStatusCode)
}

func Test_Search_HttpRequestError(t *testing.T) {
	givePayload, err := json.Marshal(Payload{})
	require.NoError(t, err)

	i := NewWithApiUrl("http://invalid-url")
	cards, httpStatusCode, err := i.Search("storeName", "storeBaseURL", givePayload)

	require.Error(t, err)
	require.Equal(t, 0, len(cards))
	require.Equal(t, 0, httpStatusCode)
}
