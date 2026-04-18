package cardscitadel

import (
	"context"
	"errors"
	"testing"

	"mtg-price-checker-sg/gateway"
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

type mockBinderposGateway struct {
	called       bool
	gotVariant   int
	gotStore     string
	gotBaseURL   string
	gotSearchURL string
	gotSearch    string
	returnCards  []gateway.Card
	returnErr    error
}

func (m *mockBinderposGateway) Scrap(ctx context.Context, scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	m.called = true
	m.gotVariant = scrapVariant
	m.gotStore = storeName
	m.gotBaseURL = baseUrl
	m.gotSearchURL = searchUrl
	m.gotSearch = searchStr
	return m.returnCards, m.returnErr
}

func TestSearchUsesBinderposGateway(t *testing.T) {
	mockGateway := &mockBinderposGateway{
		returnCards: []gateway.Card{
			{
				Name:    "Abrade",
				Source:  StoreName,
				Url:     StoreBaseURL + "/products/abrade",
				Img:     "https://example.com/abrade.png",
				Price:   1.23,
				InStock: true,
			},
		},
	}
	s := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: mockGateway,
	}

	result, err := s.Search(context.Background(), "Abrade")
	require.NoError(t, err)
	require.True(t, mockGateway.called)
	require.Equal(t, 1, mockGateway.gotVariant)
	require.Equal(t, StoreName, mockGateway.gotStore)
	require.Equal(t, StoreBaseURL, mockGateway.gotBaseURL)
	require.Equal(t, StoreSearchURL, mockGateway.gotSearchURL)
	require.Equal(t, "Abrade", mockGateway.gotSearch)
	require.Len(t, result, 1)
	require.Equal(t, "Abrade", result[0].Name)
	require.Equal(t, StoreName, result[0].Source)
	require.Contains(t, result[0].Url, StoreBaseURL+"/products/")
}

func TestSearchPropagatesBinderposError(t *testing.T) {
	mockGateway := &mockBinderposGateway{returnErr: errors.New("upstream failure")}
	s := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: mockGateway,
	}

	result, err := s.Search(context.Background(), "Abrade")
	require.Error(t, err)
	require.EqualError(t, err, "upstream failure")
	require.Nil(t, result)
}
