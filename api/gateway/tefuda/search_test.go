package tefuda

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search("sol ring")
	require.NoError(t, err)
	require.True(t, len(result) > 0)
}

func Test_scrap(t *testing.T) {
	result, err := scrap(Store{
		Name:      StoreName,
		BaseUrl:   StoreBaseURL,
		SearchUrl: StoreSearchURL,
	}, "sol ring")
	require.NoError(t, err)
	require.True(t, len(result) > 0)
}
