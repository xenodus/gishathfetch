package tefuda

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
		returnCards: []gateway.Card{{Name: "Smothering Tithe", InStock: true, Price: 22.0, Source: StoreName}},
	}
	store := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: mockGateway,
	}

	result, err := store.Search(context.Background(), "smothering tithe")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, 4, mockGateway.gotVariant)
	require.Equal(t, StoreName, mockGateway.gotStoreName)
	require.Equal(t, StoreBaseURL, mockGateway.gotBaseURL)
	require.Equal(t, StoreSearchURL, mockGateway.gotSearchURL)
	require.Equal(t, "smothering tithe", mockGateway.gotSearchStr)
}

func TestSearchPropagatesBinderposError(t *testing.T) {
	store := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: &mockBinderposGateway{returnErr: errors.New("upstream error")},
	}

	result, err := store.Search(context.Background(), "smothering tithe")
	require.Error(t, err)
	require.Nil(t, result)
}
