package cardsandcollection

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search("counterspell")
	require.NoError(t, err)
	require.True(t, len(result) > 0)
}
