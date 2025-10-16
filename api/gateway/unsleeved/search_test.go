package unsleeved

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search("ring")
	require.NoError(t, err)
	require.NotEmpty(t, result)

	for _, card := range result {
		if card.InStock {
			require.NotEmpty(t, card.Name)
			require.NotEmpty(t, card.Source)
			require.NotEmpty(t, card.Url)
			require.NotEmpty(t, card.Img)
			require.NotEmpty(t, card.Price)
			require.Contains(t, card.Url, fmt.Sprintf("%s/product/", StoreBaseURL))
			log.Println(card.Url)
		}
	}
}
