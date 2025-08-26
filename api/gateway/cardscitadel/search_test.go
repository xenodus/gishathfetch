package cardscitadel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mtg-price-checker-sg/gateway/binderpos"

	"github.com/stretchr/testify/require"
)

func Test_NewLGS(t *testing.T) {
	require.Equal(t, Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: binderpos.New(),
	}, NewLGS())
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search("Abrade")
	require.NoError(t, err)
	require.True(t, len(result) > 0)

	for _, card := range result {
		if card.InStock {
			require.NotEmpty(t, card.Name)
			require.NotEmpty(t, card.Source)
			require.NotEmpty(t, card.Url)
			require.NotEmpty(t, card.Img)
			require.NotEmpty(t, card.Price)
			require.Contains(t, card.Url, StoreBaseURL+"/products/")
		}
	}
}

func Test_Scrap(t *testing.T) {
	// Mock the Binderpos API to return a 400 error
	mockBinderposSearch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{}`))
	}))
	defer mockBinderposSearch.Close()

	// Init BinderposGwy with mocked API URL
	s := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: binderpos.NewWithApiUrl(mockBinderposSearch.URL),
	}
	result, err := s.Search("Abrade")
	require.NoError(t, err)
	require.True(t, len(result) > 0)

	for _, card := range result {
		if card.InStock {
			require.NotEmpty(t, card.Name)
			require.NotEmpty(t, card.Source)
			require.NotEmpty(t, card.Url)
			require.NotEmpty(t, card.Img)
			require.NotEmpty(t, card.Price)
			require.Contains(t, card.Url, StoreBaseURL+"/products/")
		}
	}
}
