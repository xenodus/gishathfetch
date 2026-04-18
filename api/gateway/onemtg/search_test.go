package onemtg

import (
	"context"
	"errors"
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

func TestSearchUsesBinderposGateway(t *testing.T) {
	mockGateway := &mockBinderposGateway{
		returnCards: []gateway.Card{{Name: "Abrade", InStock: true, Price: 2.5, Source: StoreName}},
	}
	s := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: mockGateway,
	}
	result, err := s.Search(context.Background(), "Abrade")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, 2, mockGateway.gotVariant)
	require.Equal(t, StoreName, mockGateway.gotStoreName)
	require.Equal(t, StoreBaseURL, mockGateway.gotBaseURL)
	require.Equal(t, StoreSearchURL, mockGateway.gotSearchURL)
	require.Equal(t, "Abrade", mockGateway.gotSearchStr)
}

func TestSearchPropagatesBinderposError(t *testing.T) {
	s := Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: &mockBinderposGateway{returnErr: errors.New("upstream error")},
	}
	result, err := s.Search(context.Background(), "Abrade")
	require.Error(t, err)
	require.Nil(t, result)
}
