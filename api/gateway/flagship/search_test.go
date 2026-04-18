package flagship

import (
	"context"
	"errors"
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

type mockBinderposGateway struct {
	gotVariant   int
	gotStoreName string
	gotBaseURL   string
	gotSearchURL string
	gotSearchStr string
	returnCards  []gateway.Card
	returnErr    error
}

func (m *mockBinderposGateway) Scrap(ctx context.Context, scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	m.gotVariant = scrapVariant
	m.gotStoreName = storeName
	m.gotBaseURL = baseUrl
	m.gotSearchURL = searchUrl
	m.gotSearchStr = searchStr
	return m.returnCards, m.returnErr
}

func TestSearchUsesBinderposGateway(t *testing.T) {
	mockGateway := &mockBinderposGateway{
		returnCards: []gateway.Card{{Name: "Abrade", InStock: true, Price: 1.2, Source: StoreName}},
	}
	store := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: mockGateway,
	}

	result, err := store.Search(context.Background(), "Abrade")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, 2, mockGateway.gotVariant)
	require.Equal(t, StoreName, mockGateway.gotStoreName)
	require.Equal(t, StoreBaseURL, mockGateway.gotBaseURL)
	require.Equal(t, StoreSearchURL, mockGateway.gotSearchURL)
	require.Equal(t, "Abrade", mockGateway.gotSearchStr)
}

func TestSearchPropagatesBinderposError(t *testing.T) {
	store := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: &mockBinderposGateway{returnErr: errors.New("upstream error")},
	}

	result, err := store.Search(context.Background(), "Abrade")
	require.Error(t, err)
	require.Nil(t, result)
}
